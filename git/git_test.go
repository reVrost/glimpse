package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDiff(t *testing.T) {
	// Get diff for all changes
	diffs, err := GetDiff()
	
	// We expect this to not error, even if there are no changes
	assert.NoError(t, err)
	assert.NotNil(t, diffs)
}

func TestGetDiffWithFiles(t *testing.T) {
	// Get diff for specific files
	diffs, err := GetDiff("README.md")
	
	// We expect this to not error, even if the file doesn't exist or has no changes
	assert.NoError(t, err)
	assert.NotNil(t, diffs)
}