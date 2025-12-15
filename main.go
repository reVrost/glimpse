package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/revrost/glimpse/config"
	"github.com/revrost/glimpse/git"
	"github.com/revrost/glimpse/llm"
	"github.com/revrost/glimpse/logs"
	"github.com/revrost/glimpse/styles"
	"github.com/revrost/glimpse/ui"
	"github.com/revrost/glimpse/watcher"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Println(styles.CreateHeader(fmt.Sprintf("Glimpse v%s (commit: %s, built: %s)", version, commit, buildTime)))
		os.Exit(0)
	}

	fmt.Println(styles.CreateHeader("Glimpse: AI-Powered Micro-Reviewer"))
	fmt.Println(ui.Separator(60))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Failed to load config: %v", err)))
		os.Exit(1)
	}

	// Initialize components
	llmClient := llm.New(llm.Config{
		Provider:     cfg.LLM.Provider,
		Model:        cfg.LLM.Model,
		APIKey:       cfg.LLM.APIKey,
		SystemPrompt: cfg.LLM.SystemPrompt,
	})
	logTailer := logs.New(logs.Config{
		File:  cfg.Logs.File,
		Lines: cfg.Logs.Lines,
	})

	// Initialize watcher
	fileWatcher, err := watcher.New(watcher.Config{
		Watch:    cfg.Watch,
		Ignore:   cfg.Ignore,
		Debounce: cfg.GetDebounceDuration(),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Failed to create watcher: %v", err)))
		os.Exit(1)
	}
	defer fileWatcher.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println(styles.Status.Render(fmt.Sprintf("Glimpse is watching %d patterns: %v", len(cfg.Watch), cfg.Watch)))
	fmt.Println(styles.Muted.Render("Press Ctrl+C to exit or type to chat (experimental)"))

	// Start the watcher
	fileWatcher.Start()

	// Main event loop with batching
	var pendingEvents []watcher.FileEvent
	var batchTimer *time.Timer

	// Git state monitoring for staging changes
	var lastStagedState *git.StagedState
	var gitStateCheckTimer *time.Timer

	// Initialize with the current staged state
	initialStagedState, err := git.GetStagedState()
	if err != nil {
		fmt.Println(styles.CreateWarningStyle(fmt.Sprintf("Warning: Could not get initial staged state: %v", err)))
	} else {
		lastStagedState = initialStagedState
	}

	// Start with a stopped timer
	batchTimer = time.NewTimer(0)
	if !batchTimer.Stop() {
		<-batchTimer.C
	}

	// Start git state checking timer (check every 500ms for responsive detection)
	gitStateCheckInterval := 500 * time.Millisecond
	gitStateCheckTimer = time.NewTimer(gitStateCheckInterval)

	for {
		select {
		case event := <-fileWatcher.Events():
			// Add to pending batch
			pendingEvents = append(pendingEvents, event)

			// Stop existing timer if any
			if batchTimer != nil {
				batchTimer.Stop()
			}

			// Start new timer
			batchTimer = time.NewTimer(cfg.GetDebounceDuration())

		case <-batchTimer.C:
			// Process batch if we have events
			if len(pendingEvents) > 0 {
				fmt.Println(styles.CreateBatchHeader(len(pendingEvents)))
				processBatch(pendingEvents, cfg, llmClient, logTailer)
				pendingEvents = nil
			}
			// Reset timer for next batch
			batchTimer = time.NewTimer(0)
			if !batchTimer.Stop() {
				<-batchTimer.C
			}

		case <-gitStateCheckTimer.C:
			// Check for staged state changes
			currentStagedState, err := git.GetStagedState()
			if err != nil {
				fmt.Println(styles.CreateWarningStyle(fmt.Sprintf("Warning: Could not check staged state: %v", err)))
			} else if lastStagedState == nil || currentStagedState.Hash != lastStagedState.Hash {
				// Staged state has changed, trigger analysis
				fmt.Println(styles.CreateWarningStyle("🔄 Git staged state changed - triggering analysis"))
				processStagedChange(currentStagedState, cfg, llmClient, logTailer)
				lastStagedState = currentStagedState
			}
			// Reset timer for next check
			gitStateCheckTimer = time.NewTimer(gitStateCheckInterval)

		case <-sigChan:
			fmt.Println(styles.CreateWarningStyle("\nShutting down Glimpse..."))
			return
		}
	}
}

// isIgnoredFile checks if a file should be ignored based on config
func isIgnoredFile(file string, cfg *config.Config) bool {
	for _, pattern := range cfg.Ignore {
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}

// processBatch handles multiple file change events in one LLM call
func processBatch(events []watcher.FileEvent, cfg *config.Config, llmClient *llm.Client, logTailer *logs.Tailer) {
	if len(events) == 0 {
		return
	}

	// Get all changed files from git (both staged and unstaged)
	changedFiles, err := git.GetChangedFiles()
	if err != nil {
		fmt.Println(styles.CreateErrorStyle(fmt.Sprintf("Error getting changed files from git: %v", err)))
		return
	}

	// Filter out ignored files
	var filteredFiles []string
	for _, file := range changedFiles {
		if !isIgnoredFile(file, cfg) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	if len(filteredFiles) == 0 {
		fmt.Println(styles.CreateWarningStyle("No reviewable changes detected (all files ignored)"))
		return
	}

	// Get diffs for all filtered files
	allDiffs, err := git.GetDiff(filteredFiles...)
	if err != nil {
		fmt.Println(styles.CreateErrorStyle(fmt.Sprintf("Error getting git diffs: %v", err)))
		return
	}

	if len(allDiffs) == 0 {
		fmt.Println(styles.CreateWarningStyle("No actual changes detected in filtered files"))
		return
	}

	// Get recent logs
	var recentLogs string
	logContent, err := logTailer.Tail()
	if err != nil {
		fmt.Println(styles.CreateWarningStyle(fmt.Sprintf("Warning: Could not read log file: %v", err)))
		recentLogs = "No logs available"
	} else {
		recentLogs = logContent
	}

	// Build context with all diffs
	var context strings.Builder
	context.WriteString(fmt.Sprintf("=== BATCH ANALYSIS: %d files changed ===\n", len(filteredFiles)))
	for i, filePath := range filteredFiles {
		context.WriteString(fmt.Sprintf("%d. %s\n", i+1, filePath))
	}
	context.WriteString("\n")

	context.WriteString("=== ALL GIT DIFFS ===\n")
	for _, diff := range allDiffs {
		context.WriteString(fmt.Sprintf("File: %s\n", diff.FilePath))
		context.WriteString("Git Diff:\n")
		context.WriteString(diff.Content)
		context.WriteString("\n\n")
	}

	// // Always display the git diff being sent to LLM for audit/debug
	// fmt.Println(styles.DiffHeader.Render("=== Git diff being sent to LLM ==="))
	// fmt.Println(styles.CreateFileList(changedFiles))
	// for _, diff := range allDiffs {
	// 	fmt.Println(styles.CreateDiffHeader(diff.FilePath))
	// 	fmt.Println(diff.Content)
	// }
	// fmt.Println(styles.DiffHeader.Render("=== END GIT DIFF ==="))

	// Show progress indicator
	progress := ui.NewProgress(100, 100, "Analyzing changes with AI")
	progress.Update(50) // Start at 50%
	fmt.Println(progress.View())

	context.WriteString(fmt.Sprintf("Recent Runtime Logs (tail -n %d):\n", cfg.Logs.Lines))
	context.WriteString(recentLogs)

	// Send to LLM with all changes
	task := fmt.Sprintf("Review the batch of %d file changes. If the logs show errors related to this logic, highlight them immediately. Be concise.", len(changedFiles))

	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      context.String(),
		Task:         task,
	}

	fmt.Println(styles.CreateProviderInfo(cfg.LLM.Provider, cfg.LLM.Model))

	// Update progress to 100%
	progress.Update(100)
	fmt.Println(progress.View())

	respChan := llmClient.Generate(req)
	resp := <-respChan
	if resp.Error != nil {
		fmt.Println(styles.CreateErrorStyle(fmt.Sprintf("LLM error: %v", resp.Error)))
	} else {
		fmt.Println(ui.SuccessBox("AI Analysis Complete", "Review has been successfully generated"))
		fmt.Println(resp.Content)
	}
}

// processStagedChange handles git staged state changes by analyzing staged diffs
func processStagedChange(stagedState *git.StagedState, cfg *config.Config, llmClient *llm.Client, logTailer *logs.Tailer) {
	if len(stagedState.StagedFiles) == 0 {
		fmt.Println(styles.CreateWarningStyle("No staged files to analyze"))
		return
	}

	// Filter out ignored files
	var filteredFiles []string
	for _, file := range stagedState.StagedFiles {
		if !isIgnoredFile(file, cfg) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	if len(filteredFiles) == 0 {
		fmt.Println(styles.CreateWarningStyle("No reviewable staged changes (all files ignored)"))
		return
	}

	// Get staged diffs for all filtered files
	stagedDiffs, err := git.GetStagedDiff(filteredFiles...)
	if err != nil {
		fmt.Println(styles.CreateErrorStyle(fmt.Sprintf("Error getting staged git diffs: %v", err)))
		return
	}

	if len(stagedDiffs) == 0 {
		fmt.Println(styles.CreateWarningStyle("No actual staged changes detected in filtered files"))
		return
	}

	// Get recent logs
	var recentLogs string
	logContent, err := logTailer.Tail()
	if err != nil {
		fmt.Println(styles.CreateWarningStyle(fmt.Sprintf("Warning: Could not read log file: %v", err)))
		recentLogs = "No logs available"
	} else {
		recentLogs = logContent
	}

	// Build context with staged diffs
	var context strings.Builder
	context.WriteString(fmt.Sprintf("=== STAGED CHANGES ANALYSIS: %d files staged ===\n", len(filteredFiles)))
	for i, filePath := range filteredFiles {
		context.WriteString(fmt.Sprintf("%d. %s\n", i+1, filePath))
	}
	context.WriteString("\n")

	context.WriteString("=== STAGED GIT DIFFS ===\n")
	for _, diff := range stagedDiffs {
		context.WriteString(fmt.Sprintf("File: %s\n", diff.FilePath))
		context.WriteString("Staged Git Diff:\n")
		context.WriteString(diff.Content)
		context.WriteString("\n\n")
	}

	// Always display the staged git diff being sent to LLM for audit/debug
	fmt.Println(styles.DiffHeader.Render("=== Staged git diff being sent to LLM ==="))
	fmt.Println(styles.CreateFileList(stagedState.StagedFiles))
	for _, diff := range stagedDiffs {
		fmt.Println(styles.CreateDiffHeader(diff.FilePath))
		fmt.Println(diff.Content)
	}
	fmt.Println(styles.DiffHeader.Render("=== END STAGED GIT DIFF ==="))

	// Show progress indicator
	progress := ui.NewProgress(100, 100, "Analyzing staged changes with AI")
	progress.Update(50) // Start at 50%
	fmt.Println(progress.View())

	context.WriteString(fmt.Sprintf("Recent Runtime Logs (tail -n %d):\n", cfg.Logs.Lines))
	context.WriteString(recentLogs)

	// Send to LLM with staged changes
	task := fmt.Sprintf("Review the batch of %d staged file changes. Focus on the staged changes specifically. If the logs show errors related to this logic, highlight them immediately. Be concise.", len(filteredFiles))

	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      context.String(),
		Task:         task,
	}

	fmt.Println(styles.CreateProviderInfo(cfg.LLM.Provider, cfg.LLM.Model))

	// Update progress to 100%
	progress.Update(100)
	fmt.Println(progress.View())

	respChan := llmClient.Generate(req)
	resp := <-respChan
	if resp.Error != nil {
		fmt.Println(styles.CreateErrorStyle(fmt.Sprintf("LLM error: %v", resp.Error)))
	} else {
		fmt.Println(ui.SuccessBox("AI Staged Changes Analysis Complete", "Staged changes review has been successfully generated"))
		fmt.Println(resp.Content)
	}
}

// Batch test change
// Another test change
// Debug test change
// Fixed timer test
// Debug main loop test
// Simplified timer test
