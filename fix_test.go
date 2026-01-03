package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFixResponse_Yes(t *testing.T) {
	input := `NEED FIX: YES

Found critical bug in payment handler.
Fix: Add error handling to the payment processing function.`

	needFix, fixPrompt, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, review, "Found critical bug")
	assert.Contains(t, fixPrompt, "error handling")
}

func TestParseFixResponse_No(t *testing.T) {
	input := `NEED FIX: NO

Code looks good. No issues found.`

	needFix, fixPrompt, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.False(t, needFix)
	assert.Contains(t, review, "Code looks good")
	assert.Empty(t, fixPrompt)
}

func TestParseFixResponse_MissingHeader_NoIssues(t *testing.T) {
	input := `The code review is complete.
No issues found. Everything looks good.`

	needFix, _, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.False(t, needFix) // Should detect "no issues" keywords
	assert.Contains(t, review, "The code review")
}

func TestParseFixResponse_MissingHeader_HasIssues(t *testing.T) {
	input := `Found memory leak in cache handler.
Potential nil pointer dereference on line 45.`

	needFix, fixPrompt, _, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix) // Should detect issues present
	assert.NotEmpty(t, fixPrompt)
}

func TestParseFixResponse_Empty(t *testing.T) {
	needFix, fixPrompt, review, err := parseFixResponse("")

	assert.Error(t, err)
	assert.False(t, needFix)
	assert.Empty(t, fixPrompt)
	assert.Empty(t, review)
}

func TestParseFixResponse_MultipleFixes(t *testing.T) {
	input := `NEED FIX: YES

Issue 1: Missing null check
Issue 2: Unclosed file handle

Fix: Add null check before dereferencing pointer and ensure file is closed with defer.`

	needFix, fixPrompt, _, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, fixPrompt, "null check")
	assert.Contains(t, fixPrompt, "file is closed")
}

func TestParseFixResponse_WhitespaceInHeader(t *testing.T) {
	input := `  NEED FIX: YES  

Review text here.`

	needFix, _, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, review, "Review text here")
}

func TestParseFixResponse_CaseInsensitiveHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lowercase yes", "need fix: yes\nReview", true},
		{"mixed case yes", "Need Fix: YES\nReview", true},
		{"uppercase no", "NEED FIX: NO\nReview", false},
		{"lowercase no", "need fix: no\nReview", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			needFix, _, _, err := parseFixResponse(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, needFix)
		})
	}
}

func TestParseFixResponse_HeaderNotFirstLine(t *testing.T) {
	input := `

NEED FIX: YES

Review content here.`

	needFix, _, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, review, "Review content here")
}

func TestExtractFixPrompt_ExplicitFixSection(t *testing.T) {
	input := `Review content.

Fix: Add error handling for nil pointer.
Update the validate function.

More content.`

	fixPrompt := extractFixPrompt(input)

	assert.Contains(t, fixPrompt, "error handling")
	assert.Contains(t, fixPrompt, "validate function")
}

func TestExtractFixPrompt_FixInstruction(t *testing.T) {
	input := `Issues found.

Fix instruction: Implement retry logic for API calls.

More text.`

	fixPrompt := extractFixPrompt(input)

	assert.Contains(t, fixPrompt, "retry logic")
}

func TestExtractFixPrompt_LastParagraphWithKeywords(t *testing.T) {
	input := `First paragraph about code structure.
Second paragraph about naming conventions.

The last paragraph should fix the memory leak by adding proper cleanup.`

	fixPrompt := extractFixPrompt(input)

	assert.Contains(t, fixPrompt, "memory leak")
	assert.Contains(t, fixPrompt, "proper cleanup")
}

func TestExtractFixPrompt_MultipleSections(t *testing.T) {
	input := `Issue 1.

Fix: Add null check.

Issue 2.

Fix: Close file handle.

Issue 3.`

	fixPrompt := extractFixPrompt(input)

	assert.Contains(t, fixPrompt, "null check")
	assert.Contains(t, fixPrompt, "Close file handle")
}

func TestReverseStrings(t *testing.T) {
	input := []string{"a", "b", "c", "d"}
	expected := []string{"d", "c", "b", "a"}

	result := reverseStrings(input)

	assert.Equal(t, expected, result)
}
