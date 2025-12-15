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

const (
	maxBatchSize      = 100
	idleTimerDuration = time.Hour
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Println(
			styles.CreateHeader(
				fmt.Sprintf("Glimpse v%s (commit: %s, built: %s)", version, commit, buildTime),
			),
		)
		return
	}

	fmt.Println(styles.CreateHeader("Glimpse: AI-Powered Micro-Reviewer"))
	fmt.Println(ui.Separator(60))

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(err.Error()))
		os.Exit(1)
	}

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

	fileWatcher, err := watcher.New(watcher.Config{
		Watch:    cfg.Watch,
		Ignore:   cfg.Ignore,
		Debounce: cfg.GetDebounceDuration(),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(err.Error()))
		os.Exit(1)
	}
	defer fileWatcher.Close()

	done := make(chan struct{})
	batchChan := make(chan []watcher.FileEvent, 5)

	startBatcher(
		fileWatcher.Events(),
		batchChan,
		cfg.GetDebounceDuration(),
		done,
	)

	fileWatcher.Start()

	fmt.Println(
		styles.Status.Render(
			fmt.Sprintf("Watching %d patterns: %v", len(cfg.Watch), cfg.Watch),
		),
	)
	fmt.Println(styles.Muted.Render("Press Ctrl+C to exit"))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var lastStagedHash string
	gitTicker := time.NewTicker(1 * time.Second)
	defer gitTicker.Stop()

	for {
		select {
		// -- Disabled file watching for now
		// case batch := <-batchChan:
		// fmt.Println(styles.CreateBatchHeader(len(batch)))
		// fmt.Println(batch)
		// processBatch(batch, cfg, llmClient, logTailer)

		case <-gitTicker.C:
			state, err := git.GetStagedState()
			if err == nil && state.Hash != lastStagedHash {
				lastStagedHash = state.Hash
				// fmt.Println(styles.CreateBatchHeader(len(batch)))
				fmt.Println(styles.CreateGlossMessage("Git staged state changed. Reviewing..."))
				processStagedChange(state, cfg, llmClient, logTailer)
			}

		case <-sigChan:
			fmt.Println(styles.CreateWarningStyle("\nShutting down Glimpse..."))
			close(done)
			return
		}
	}
}

/* ----------------------- Helpers ----------------------- */

func isIgnoredFile(file string, cfg *config.Config) bool {
	// Always ignore .git directory changes to prevent infinite loops
	// caused by your own git polling.
	if strings.Contains(file, "/.git/") || strings.HasPrefix(file, ".git/") {
		return true
	}

	// 2. Ignore the application's own log file if it's inside the repo
	if strings.Contains(file, cfg.Logs.File) {
		return true
	}

	// 3. Check user config
	for _, pattern := range cfg.Ignore {
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}

/* ----------------------- Batching ---------------------- */

func startBatcher(
	events <-chan watcher.FileEvent,
	out chan<- []watcher.FileEvent,
	debounce time.Duration,
	done <-chan struct{},
) {
	go func() {
		var batch []watcher.FileEvent
		timer := time.NewTimer(idleTimerDuration)

		stopAndDrain := func(t *time.Timer) {
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
		}

		for {
			select {
			case <-done:
				stopAndDrain(timer)
				if len(batch) > 0 {
					out <- batch
				}
				return

			case ev := <-events:
				batch = append(batch, ev)

				if len(batch) >= maxBatchSize {
					out <- batch
					batch = nil
					stopAndDrain(timer)
					timer.Reset(idleTimerDuration)
					continue
				}

				stopAndDrain(timer)
				timer.Reset(debounce)

			case <-timer.C:
				if len(batch) > 0 {
					out <- batch
					batch = nil
				}
				timer.Reset(idleTimerDuration)
			}
		}
	}()
}

/* -------------------- Batch Processing -------------------- */

func processBatch(
	events []watcher.FileEvent,
	cfg *config.Config,
	llmClient *llm.Client,
	logTailer *logs.Tailer,
) {
	fileSet := make(map[string]struct{})
	for _, ev := range events {
		if !isIgnoredFile(ev.Path, cfg) {
			fileSet[ev.Path] = struct{}{}
		}
	}

	if len(fileSet) == 0 {
		return
	}

	var files []string
	for f := range fileSet {
		files = append(files, f)
	}

	diffs, err := git.GetDiff(files...)
	if err != nil || len(diffs) == 0 {
		return
	}

	logsText, _ := logTailer.Tail()

	var ctx strings.Builder
	ctx.WriteString("=== FILE CHANGE REVIEW ===\n")
	for _, d := range diffs {
		ctx.WriteString(fmt.Sprintf("File: %s\n%s\n\n", d.FilePath, d.Content))
	}
	ctx.WriteString("=== RUNTIME LOGS ===\n")
	ctx.WriteString(logsText)

	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      ctx.String(),
		Task:         "Review these changes and flag bugs or risks. Be concise.",
	}

	fmt.Println(styles.CreateProviderInfo(cfg.LLM.Provider, cfg.LLM.Model))
	launchLLMAsync(llmClient, req, "AI Analysis Complete")
}

/* -------------------- Staged Processing -------------------- */

func processStagedChange(
	state *git.StagedState,
	cfg *config.Config,
	llmClient *llm.Client,
	logTailer *logs.Tailer,
) {
	if len(state.StagedFiles) == 0 {
		return
	}

	var files []string
	for _, f := range state.StagedFiles {
		if !isIgnoredFile(f, cfg) {
			files = append(files, f)
		}
	}

	diffs, err := git.GetStagedDiff(files...)
	if err != nil || len(diffs) == 0 {
		return
	}

	logsText, _ := logTailer.Tail()

	var ctx strings.Builder
	ctx.WriteString("=== STAGED CHANGE REVIEW ===\n")
	for _, d := range diffs {
		ctx.WriteString(fmt.Sprintf("File: %s\n%s\n\n", d.FilePath, d.Content))
	}
	ctx.WriteString("=== RUNTIME LOGS ===\n")
	ctx.WriteString(logsText)

	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      ctx.String(),
		Task:         "Review staged changes only. Flag bugs or risks. Be concise.",
	}

	fmt.Println(styles.CreateProviderInfo(cfg.LLM.Provider, cfg.LLM.Model))
	launchLLMAsync(llmClient, req, "AI Staged Review Complete")
}

/* ---------------------- LLM Runner ---------------------- */

func launchLLMAsync(
	client *llm.Client,
	req llm.GenerateRequest,
	title string,
) {
	go func() {
		resp := <-client.Generate(req)
		if resp.Error != nil {
			fmt.Println(styles.CreateErrorStyle(resp.Error.Error()))
			return
		}
		fmt.Println(ui.SuccessBox(title, "Review generated successfully"))
		fmt.Println(resp.Content)
	}()
}
