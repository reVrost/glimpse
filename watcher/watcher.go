package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/revrost/glimpse/styles"
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
	fmt.Println(styles.Info.Render(fmt.Sprintf("Processing %d watch patterns", len(config.Watch))))
	for _, pattern := range config.Watch {
		fmt.Println(styles.Muted.Render(fmt.Sprintf("Pattern: %s", pattern)))
		
		// Check if pattern contains ** (recursive)
		if strings.Contains(pattern, "**") {
			// Handle recursive pattern
			baseDir := strings.Split(pattern, "**")[0]
			baseDir = strings.TrimSuffix(baseDir, "/")
			if baseDir == "" {
				baseDir = "."
			}
			
			fmt.Println(styles.Text.Render(fmt.Sprintf("Adding directory for recursive pattern: %s", baseDir)))
			if err := w.watcher.Add(baseDir); err != nil {
				fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Failed to watch directory %s: %v", baseDir, err)))
				continue
			}
			addedDirs[baseDir] = true
		} else {
			// Handle standard glob pattern
			matches, err := filepath.Glob(pattern)
			if err != nil {
				fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Glob error for pattern %s: %v", pattern, err)))
				continue
			}
			fmt.Println(styles.Muted.Render(fmt.Sprintf("Glob matches for %s: %v", pattern, matches)))
			
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
						fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Failed to watch directory %s: %v", dir, err)))
						continue
					}
					fmt.Println(styles.Info.Render(fmt.Sprintf("Watching directory: %s", dir)))
					addedDirs[dir] = true
				} else {
					fmt.Println(styles.Muted.Render(fmt.Sprintf("Directory %s already being watched", dir)))
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

// Start begins watching for file changes (no debouncing - handled in main loop)
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				
				// Skip if event should be ignored
				if w.shouldIgnore(event.Name) {
					continue
				}

				// Normalize path to handle editor temporary files
				normalizedPath := w.normalizePath(event.Name)

				// Send event immediately (batching handled in main loop)
				w.events <- FileEvent{Path: normalizedPath}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Watcher error: %v", err)))
			}
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