package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/revrost/glimpse/config"
	"github.com/revrost/glimpse/git"
	"github.com/revrost/glimpse/llm"
	"github.com/revrost/glimpse/logs"
	"github.com/revrost/glimpse/watcher"
)

func main() {
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
	
	// Main event loop
	for {
		select {
		case event := <-fileWatcher.Events():
			// Process file change
			fmt.Printf("\n--- Reviewing: %s ---\n", event.Path)
			processEvent(event, cfg, llmClient, logTailer)
			
		case <-sigChan:
			fmt.Println("\nShutting down Glimpse...")
			return
		}
	}
}

// processEvent handles a file change event
func processEvent(event watcher.FileEvent, cfg *config.Config, llmClient *llm.Client, logTailer *logs.Tailer) {
	// Get git diff
	diffs, err := git.GetDiff(event.Path)
	if err != nil {
		fmt.Printf("Error getting git diff: %v\n", err)
		return
	}
	
	if len(diffs) == 0 {
		fmt.Println("No changes detected")
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
	
	// Build context for LLM
	var context strings.Builder
	for _, diff := range diffs {
		context.WriteString(fmt.Sprintf("File: %s\n", diff.FilePath))
		context.WriteString("Git Diff:\n")
		context.WriteString(diff.Content)
		context.WriteString("\n\n")
	}
	
	context.WriteString(fmt.Sprintf("Recent Runtime Logs (tail -n %d):\n", cfg.Logs.Lines))
	context.WriteString(recentLogs)
	
	// Send to LLM
	task := "Review the diff. If the logs show errors related to this logic, highlight them immediately. Be concise."
	
	req := llm.GenerateRequest{
		SystemPrompt: cfg.LLM.SystemPrompt,
		Context:      context.String(),
		Task:         task,
	}
	
	fmt.Printf("Analyzing with %s (%s)...\n", cfg.LLM.Provider, cfg.LLM.Model)
	
	respChan := llmClient.Generate(req)
	resp := <-respChan
	if resp.Error != nil {
		fmt.Printf("LLM error: %v\n", resp.Error)
	} else {
		fmt.Println(resp.Content)
	}
}