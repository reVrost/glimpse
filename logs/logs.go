package logs

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds the log configuration
type Config struct {
	File  string
	Lines int
}

// Tailer represents a log tailing instance
type Tailer struct {
	config Config
}

// New creates a new log tailer instance
func New(config Config) *Tailer {
	return &Tailer{
		config: config,
	}
}

// Tail returns the last N lines from the log file
func (t *Tailer) Tail() (string, error) {
	file, err := os.Open(t.config.File)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	
	// Read all lines first (for simplicity)
	// In production, would use a more efficient approach for large files
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading log file: %w", err)
	}
	
	// Get the last N lines
	start := 0
	if len(lines) > t.config.Lines {
		start = len(lines) - t.config.Lines
	}
	
	recentLines := lines[start:]
	return strings.Join(recentLines, "\n"), nil
}