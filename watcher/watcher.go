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
	var timer *time.Timer
	debounceChan := make(chan string, 1)

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

				fmt.Printf("Processing event: %s\n", event.Name)

				// Debounce logic
				if timer != nil {
					timer.Stop()
				}
				
				timer = time.AfterFunc(w.config.Debounce, func() {
					select {
					case debounceChan <- event.Name:
						fmt.Printf("Debounced event sent: %s\n", event.Name)
					default:
						// Channel already has a pending event
						fmt.Printf("Debounce channel full for: %s\n", event.Name)
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