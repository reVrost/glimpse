package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStagedState(t *testing.T) {
	// Get staged state
	stagedState, err := GetStagedState()
	
	// We expect this to not error
	assert.NoError(t, err)
	assert.NotNil(t, stagedState)
	assert.NotNil(t, stagedState.StagedFiles)
	assert.NotEmpty(t, stagedState.Hash)
}

func TestGetStagedDiff(t *testing.T) {
	// Get staged diff
	diffs, err := GetStagedDiff()
	
	// We expect this to not error
	assert.NoError(t, err)
	assert.NotNil(t, diffs)
}