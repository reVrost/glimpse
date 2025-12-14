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
		fmt.Printf("Glimpse v%s (commit: %s, built: %s)\n", version, commit, buildTime)
		os.Exit(0)
	}

	fmt.Println("Glimpse: AI-Powered Micro-Reviewer")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
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
		Watch:   cfg.Watch,
		Ignore:  cfg.Ignore,
		Debounce: cfg.GetDebounceDuration(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create watcher: %v\n", err)
		os.Exit(1)
	}
	defer fileWatcher.Close()
	
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	fmt.Printf("Glimpse is watching %d patterns: %v\n", len(cfg.Watch), cfg.Watch)
	fmt.Println("Press Ctrl+C to exit or type to chat (experimental)")
	
	// Start the watcher
	fileWatcher.Start()
	
	// Main event loop with batching
	var pendingEvents []watcher.FileEvent
	var batchTimer *time.Timer
	
	// Start with a stopped timer
	batchTimer = time.NewTimer(0)
	if !batchTimer.Stop() {
		<-batchTimer.C
	}
	
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
				fmt.Printf("\n--- Processing batch of %d changes ---\n", len(pendingEvents))
				processBatch(pendingEvents, cfg, llmClient, logTailer)
				pendingEvents = nil
			}
			// Reset timer for next batch
			batchTimer = time.NewTimer(0)
			if !batchTimer.Stop() {
				<-batchTimer.C
			}
			
		case <-sigChan:
			fmt.Println("\nShutting down Glimpse...")
			return
		}
	}
}

// processBatch handles multiple file change events in one LLM call
func processBatch(events []watcher.FileEvent, cfg *config.Config, llmClient *llm.Client, logTailer *logs.Tailer) {
	if len(events) == 0 {
		return
	}
	
	// Collect all diffs from all files
	var allDiffs []git.Diff
	var changedFiles []string
	
	for _, event := range events {
		diffs, err := git.GetDiff(event.Path)
		if err != nil {
			fmt.Printf("Error getting git diff for %s: %v\n", event.Path, err)
			continue
		}
		
		if len(diffs) > 0 {
			allDiffs = append(allDiffs, diffs...)
			changedFiles = append(changedFiles, event.Path)
		}
	}
	
	if len(allDiffs) == 0 {
		fmt.Println("No changes detected in batch")
		return
	}
	
	// Get recent logs
	var recentLogs string
	logContent, err := logTailer.Tail()
	if err != nil {
		fmt.Printf("Warning: Could not read log file: %v\n", err)
		recentLogs = "No logs available"
	} else {
		recentLogs = logContent
	}
	
	// Build context with all diffs
	var context strings.Builder
	context.WriteString(fmt.Sprintf("=== BATCH ANALYSIS: %d files changed ===\n", len(changedFiles)))
	for i, filePath := range changedFiles {
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
	
	// Always display the git diff being sent to LLM for audit/debug
	fmt.Println("\n=== Git diff being sent to LLM ===")
	fmt.Printf("Files changed: %v\n", changedFiles)
	for _, diff := range allDiffs {
		fmt.Printf("\n--- File: %s ---\n", diff.FilePath)
		fmt.Println(diff.Content)
	}
	fmt.Println("=== END GIT DIFF ===")
	
	context.WriteString(fmt.Sprintf("Recent Runtime Logs (tail -n %d):\n", cfg.Logs.Lines))
	context.WriteString(recentLogs)
	
	// Send to LLM with all changes
	task := fmt.Sprintf("Review the batch of %d file changes. If the logs show errors related to this logic, highlight them immediately. Be concise.", len(changedFiles))
	
	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      context.String(),
		Task:         task,
	}
	
	fmt.Printf("Analyzing batch of %d changes with %s (%s)...\n", len(changedFiles), cfg.LLM.Provider, cfg.LLM.Model)
	
	respChan := llmClient.Generate(req)
	resp := <-respChan
	if resp.Error != nil {
		fmt.Printf("LLM error: %v\n", resp.Error)
	} else {
		fmt.Println(resp.Content)
	}
}
// Batch test change
// Another test change
// Debug test change
// Fixed timer test
// Debug main loop test
// Simplified timer test
