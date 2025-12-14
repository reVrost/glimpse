package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := Config{
		Watch:   []string{"./test/*.go"},
		Ignore:  []string{"*_test.go"},
		Debounce: 100 * time.Millisecond,
	}
	
	w, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.Equal(t, config, w.config)
	assert.NotNil(t, w.events)
	
	// Clean up
	w.Close()
}

func TestWatcherIgnoresPatterns(t *testing.T) {
	w := &Watcher{
		config: Config{
			Ignore: []string{"*_test.go", "*.tmp"},
		},
	}
	
	assert.True(t, w.shouldIgnore("example_test.go"))
	assert.True(t, w.shouldIgnore("temp.tmp"))
	assert.False(t, w.shouldIgnore("example.go"))
	assert.False(t, w.shouldIgnore("main.go"))
}

func TestWatcherDetectsFileChanges(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	config := Config{
		Watch:   []string{tmpDir + "/**"}, // Use recursive pattern
		Debounce: 50 * time.Millisecond,
	}
	
	w, err := New(config)
	assert.NoError(t, err)
	defer w.Close()
	
	w.Start()
	
	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("initial"), 0644)
	assert.NoError(t, err)
	
	// Give time for file system events to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Modify the file
	err = os.WriteFile(testFile, []byte("modified"), 0644)
	assert.NoError(t, err)
	
	// Wait for file system events
	select {
	case event := <-w.Events():
		assert.Contains(t, event.Path, "test.txt")
	case <-time.After(2 * time.Second):
		t.Fatal("Expected file change event but got none")
	}
}