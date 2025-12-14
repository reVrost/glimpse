package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	w := &Watcher{}
	
	tests := []struct {
		input    string
		expected string
	}{
		{"debug.log", "debug.log"},
		{"debug.log~", "debug.log"},
		{"main.go.swp", "main.go"},
		{"config.tmp", "config"},
		{".#lock.go", "lock.go"},
		{"#auto-save.go#", "auto-save.go"},
		{"/path/to/debug.log~", "/path/to/debug.log"},
	}
	
	for _, test := range tests {
		result := w.normalizePath(test.input)
		assert.Equal(t, test.expected, result, "Failed for input: %s", test.input)
	}
}

func TestDebouncingWithMultipleFiles(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	config := Config{
		Watch:   []string{tmpDir + "/**"},
		Debounce: 100 * time.Millisecond,
	}
	
	w, err := New(config)
	assert.NoError(t, err)
	defer w.Close()
	
	w.Start()
	
	// Create test files
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")
	
	// Rapidly modify file1 multiple times
	for i := 0; i < 5; i++ {
		err = os.WriteFile(file1, []byte("change"+string(rune(i))), 0644)
		assert.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Faster than debounce
	}
	
	// Rapidly modify file2 multiple times
	for i := 0; i < 5; i++ {
		err = os.WriteFile(file2, []byte("change"+string(rune(i))), 0644)
		assert.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Faster than debounce
	}
	
	// Wait for debounce to settle
	time.Sleep(200 * time.Millisecond)
	
	// We should only receive one event per file (normalized path)
	eventCount := 0
	receivedFiles := make(map[string]bool)
	
	// Collect events with timeout
	timeout := time.After(2 * time.Second)
	
	for eventCount < 2 {
		select {
		case event := <-w.Events():
			receivedFiles[event.Path] = true
			eventCount++
			t.Logf("Received event for: %s (total: %d)", event.Path, eventCount)
		case <-timeout:
			t.Fatal("Timeout waiting for events")
		}
	}
	
	// Verify we got one event per normalized file
	assert.Contains(t, receivedFiles, file1)
	assert.Contains(t, receivedFiles, file2)
	assert.Equal(t, 2, len(receivedFiles))
}

func TestEditorTemporaryFileDebouncing(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	config := Config{
		Watch:   []string{tmpDir + "/**"},
		Debounce: 100 * time.Millisecond,
	}
	
	w, err := New(config)
	assert.NoError(t, err)
	defer w.Close()
	
	w.Start()
	
	// Create main file
	mainFile := filepath.Join(tmpDir, "debug.log")
	err = os.WriteFile(mainFile, []byte("initial"), 0644)
	assert.NoError(t, err)
	
	// Simulate editor creating temporary file
	tempFile := filepath.Join(tmpDir, "debug.log~")
	err = os.WriteFile(tempFile, []byte("temp content"), 0644)
	assert.NoError(t, err)
	
	// Wait a bit and modify main file
	time.Sleep(50 * time.Millisecond)
	err = os.WriteFile(mainFile, []byte("modified"), 0644)
	assert.NoError(t, err)
	
	// Wait for debounce to settle
	time.Sleep(200 * time.Millisecond)
	
	// Should only receive one event for debug.log (normalized from both files)
	select {
	case event := <-w.Events():
		assert.Equal(t, mainFile, event.Path)
		t.Logf("Received normalized event for: %s", event.Path)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for normalized event")
	}
}