package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFixResponse_Yes(t *testing.T) {
	input := `NEED FIX: YES

Found critical bug in payment handler.
Fix: Add error handling to the payment processing function.`

	needFix, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, review, "Found critical bug")
}

func TestParseFixResponse_No(t *testing.T) {
	input := `NEED FIX: NO

Code looks good. No issues found.`

	needFix, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.False(t, needFix)
	assert.Contains(t, review, "Code looks good")
}

func TestParseFixResponse_MissingHeader_NoIssues(t *testing.T) {
	input := `The code review is complete.
No issues found. Everything looks good.`

	needFix, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.False(t, needFix) // Should detect "no issues" keywords
	assert.Contains(t, review, "The code review")
}

func TestParseFixResponse_MissingHeader_HasIssues(t *testing.T) {
	input := `Found memory leak in cache handler.
Potential nil pointer dereference on line 45.`

	needFix, _, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix) // Should detect issues present
}

func TestParseFixResponse_Empty(t *testing.T) {
	needFix, review, err := parseFixResponse("")

	assert.Error(t, err)
	assert.False(t, needFix)
	assert.Empty(t, review)
}

func TestParseFixResponse_MultipleFixes(t *testing.T) {
	input := `NEED FIX: YES

Issue 1: Missing null check
Issue 2: Unclosed file handle

Fix: Add null check before dereferencing pointer and ensure file is closed with defer.`

	needFix, _, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
}

func TestParseFixResponse_WhitespaceInHeader(t *testing.T) {
	input := `  NEED FIX: YES

Review text here.`

	needFix, review, err := parseFixResponse(input)

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
			needFix, _, err := parseFixResponse(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, needFix)
		})
	}
}

func TestParseFixResponse_HeaderNotFirstLine(t *testing.T) {
	input := `

NEED FIX: YES

Review content here.`

	needFix, review, err := parseFixResponse(input)

	assert.NoError(t, err)
	assert.True(t, needFix)
	assert.Contains(t, review, "Review content here")
}

func TestReverseStrings(t *testing.T) {
	input := []string{"a", "b", "c", "d"}
	expected := []string{"d", "c", "b", "a"}

	result := reverseStrings(input)

	assert.Equal(t, expected, result)
}
