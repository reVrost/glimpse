package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Config holds the watcher configuration
type Config struct {
	Watch   []string
	Ignore  []string
	Debounce time.Duration
}

// Watcher monitors filesystem changes
type Watcher struct {
	config  Config
	watcher *fsnotify.Watcher
	events  chan FileEvent
}

// FileEvent represents a file change event
type FileEvent struct {
	Path string
}

// New creates a new watcher instance
func New(config Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		config:  config,
		watcher: fsWatcher,
		events:  make(chan FileEvent, 100),
	}

	// Add watch patterns - watch directories, not files
	addedDirs := make(map[string]bool)
	fmt.Printf("Processing %d watch patterns\n", len(config.Watch))
	for _, pattern := range config.Watch {
		fmt.Printf("Pattern: %s\n", pattern)
		
		// Check if pattern contains ** (recursive)
		if strings.Contains(pattern, "**") {
			// Handle recursive pattern
			baseDir := strings.Split(pattern, "**")[0]
			baseDir = strings.TrimSuffix(baseDir, "/")
			if baseDir == "" {
				baseDir = "."
			}
			
			fmt.Printf("Adding directory for recursive pattern: %s\n", baseDir)
			if err := w.watcher.Add(baseDir); err != nil {
				fmt.Printf("Failed to watch directory %s: %v\n", baseDir, err)
				continue
			}
			addedDirs[baseDir] = true
		} else {
			// Handle standard glob pattern
			matches, err := filepath.Glob(pattern)
			if err != nil {
				fmt.Printf("Glob error for pattern %s: %v\n", pattern, err)
				continue
			}
			fmt.Printf("Glob matches for %s: %v\n", pattern, matches)
			
			for _, match := range matches {
				// Get the directory to watch
				dir := match
				info, err := os.Stat(match)
				if err == nil && !info.IsDir() {
					dir = filepath.Dir(match)
				}
				
				// Add directory if not already added
				if !addedDirs[dir] {
					if err := w.watcher.Add(dir); err != nil {
						fmt.Printf("Failed to watch directory %s: %v\n", dir, err)
						continue
					}
					fmt.Printf("Watching directory: %s\n", dir)
					addedDirs[dir] = true
				} else {
					fmt.Printf("Directory %s already being watched\n", dir)
				}
			}
		}
	}

	return w, nil
}

// Events returns the channel of file events
func (w *Watcher) Events() <-chan FileEvent {
	return w.events
}

// Start begins watching for file changes with debouncing
func (w *Watcher) Start() {
	debounceTimers := make(map[string]*time.Timer)
	debounceChan := make(chan string, 100)

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				
				fmt.Printf("FS Event: %s\n", event.Name)
				
				// Skip if event should be ignored
				if w.shouldIgnore(event.Name) {
					fmt.Printf("Ignoring event: %s\n", event.Name)
					continue
				}

				// Normalize path to handle editor temporary files
				normalizedPath := w.normalizePath(event.Name)
				fmt.Printf("Processing event: %s (normalized: %s)\n", event.Name, normalizedPath)

				// Cancel existing timer for this file if any
				if timer, exists := debounceTimers[normalizedPath]; exists {
					timer.Stop()
				}
				
				// Create new timer for this file
				debounceTimers[normalizedPath] = time.AfterFunc(w.config.Debounce, func() {
					select {
					case debounceChan <- normalizedPath:
						fmt.Printf("Debounced event sent: %s\n", normalizedPath)
						// Clean up timer reference
						delete(debounceTimers, normalizedPath)
					default:
						fmt.Printf("Debounce channel full for: %s\n", normalizedPath)
					}
				})

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	// Process debounced events
	go func() {
		for path := range debounceChan {
			w.events <- FileEvent{Path: path}
		}
	}()
}

// normalizePath handles editor temporary files by collapsing them to their source
func (w *Watcher) normalizePath(path string) string {
	// Handle common editor temporary file patterns
	if strings.HasSuffix(path, "~") {
		// Vim/emacs backup files
		return strings.TrimSuffix(path, "~")
	}
	if strings.HasSuffix(path, ".swp") {
		// Vim swap files
		return strings.TrimSuffix(path, ".swp")
	}
	if strings.HasSuffix(path, ".tmp") {
		// Generic temporary files
		return strings.TrimSuffix(path, ".tmp")
	}
	if strings.HasPrefix(filepath.Base(path), ".#") {
		// Emacs lock files
		return filepath.Join(filepath.Dir(path), strings.TrimPrefix(filepath.Base(path), ".#"))
	}
	if strings.HasPrefix(filepath.Base(path), "#") && strings.HasSuffix(filepath.Base(path), "#") {
		// Emacs auto-save files
		basename := strings.TrimPrefix(filepath.Base(path), "#")
		basename = strings.TrimSuffix(basename, "#")
		return filepath.Join(filepath.Dir(path), basename)
	}
	
	return path
}

// shouldIgnore checks if a file should be ignored
func (w *Watcher) shouldIgnore(path string) bool {
	for _, pattern := range w.config.Ignore {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}