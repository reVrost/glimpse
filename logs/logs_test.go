package logs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTail(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Write test log lines
	logContent := `Line 1
Line 2
Line 3
Line 4
Line 5`
	
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)
	
	// Create a tailer
	tailer := New(Config{
		File:  logFile,
		Lines: 3,
	})
	
	// Test tailing
	result, err := tailer.Tail()
	assert.NoError(t, err)
	
	// Should return the last 3 lines
	expected := `Line 3
Line 4
Line 5`
	assert.Equal(t, expected, result)
}

func TestTailNonExistentFile(t *testing.T) {
	tailer := New(Config{
		File:  "nonexistent.log",
		Lines: 10,
	})
	
	// Test tailing non-existent file
	_, err := tailer.Tail()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open log file")
}

func TestTailZeroLines(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Write test log lines
	logContent := `Line 1
Line 2
Line 3`
	
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)
	
	// Create a tailer with 0 lines
	tailer := New(Config{
		File:  logFile,
		Lines: 0,
	})
	
	// Test tailing
	result, err := tailer.Tail()
	assert.NoError(t, err)
	assert.Equal(t, "", result) // No lines should be returned
}

func TestTailMoreLinesThanExist(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Write test log lines
	logContent := `Line 1
Line 2`
	
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)
	
	// Create a tailer requesting more lines than exist
	tailer := New(Config{
		File:  logFile,
		Lines: 5,
	})
	
	// Test tailing
	result, err := tailer.Tail()
	assert.NoError(t, err)
	
	// Should return all lines
	assert.Equal(t, logContent, result)
}