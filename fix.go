package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/revrost/glimpse/styles"
)

const (
	crushTimeout = 5 * time.Minute
)

// parseFixResponse parses the LLM response to extract fix decision and review
func parseFixResponse(content string) (needFix bool, review string, err error) {
	if strings.TrimSpace(content) == "" {
		return false, "", fmt.Errorf("empty response")
	}

	lines := strings.Split(content, "\n")

	// Find first non-empty line for header detection
	var headerLine int
	var header string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			headerLine = i
			header = trimmed
			break
		}
	}

	// Check for explicit header (case-insensitive)
	headerUpper := strings.ToUpper(header)
	if strings.HasPrefix(headerUpper, "NEED FIX: YES") {
		needFix = true
		// Review starts after header line
		review = strings.Join(lines[headerLine+1:], "\n")
	} else if strings.HasPrefix(headerUpper, "NEED FIX: NO") {
		needFix = false
		review = strings.Join(lines[headerLine+1:], "\n")
	} else {
		// Fallback: keyword detection
		contentLower := strings.ToLower(content)
		noIssuesKeywords := []string{"no issues", "looks good", "everything is fine", "no problems", "all good"}

		// Check if any "no issues" keywords are present
		hasNoIssues := false
		for _, keyword := range noIssuesKeywords {
			if strings.Contains(contentLower, keyword) {
				hasNoIssues = true
				break
			}
		}

		needFix = !hasNoIssues
		review = content // No header, use entire content as review
	}

	// Clean up review: remove leading/trailing whitespace
	review = strings.TrimSpace(review)

	return needFix, review, nil
}

// reverseStrings reverses a slice of strings
func reverseStrings(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// runCrushFix executes crush with the review and streams output
func runCrushFix(review string) error {
	// Check if crush is installed
	_, err := exec.LookPath("crush")
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(
			"crush not found. Install with: go install github.com/charmbracelet/crush@latest",
		))
		return fmt.Errorf("crush binary not found")
	}

	// Prepend simple instruction to the review
	prompt := "Fix all critical reviews mentioned in above:\n\n" + review

	// Truncate prompt if too long (avoid command line limits)
	const maxPromptLength = 10000
	if len(prompt) > maxPromptLength {
		fmt.Fprintln(os.Stderr, styles.CreateWarningStyle(
			fmt.Sprintf("Review truncated from %d to %d characters", len(prompt), maxPromptLength),
		))
		prompt = prompt[:maxPromptLength]
	}

	fmt.Println(styles.CreateHeader("--- RUNNING CRUSH TO FIX ---"))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), crushTimeout)
	defer cancel()

	// Run crush command
	cmd := exec.CommandContext(ctx, "crush", "run", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Stream output after command completes
	fmt.Println(stdout.String())
	if stderr.String() != "" {
		fmt.Fprintln(os.Stderr, stderr.String())
	}

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprintln(os.Stderr, styles.CreateWarningStyle(
			"crush execution timed out after 5 minutes",
		))
		return fmt.Errorf("crush timeout")
	}

	// Handle execution error
	if err != nil {
		fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(
			fmt.Sprintf("crush execution failed: %v", err),
		))
		return fmt.Errorf("crush failed: %w", err)
	}

	return nil
}
