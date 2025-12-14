package git

import (
	"bytes"
	"os/exec"
)

// Diff represents a git diff for a specific file
type Diff struct {
	FilePath string
	Content  string
}

// GetDiff returns the git diff for the specified files or all changes if no files specified
func GetDiff(files ...string) ([]Diff, error) {
	var diffs []Diff
	
	// If no files specified, get diff for all changed files
	if len(files) == 0 {
		cmd := exec.Command("git", "diff", "--name-only", "HEAD")
		var out bytes.Buffer
		cmd.Stdout = &out
		
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		
		// Parse output to get file list
		// This is a simplified approach - in production would use proper parsing
		// For now, we'll get a general diff
		return getAllDiffs()
	}
	
	// Get diff for specific files
	for _, file := range files {
		cmd := exec.Command("git", "diff", "--unified=0", "HEAD", "--", file)
		var out bytes.Buffer
		cmd.Stdout = &out
		
		if err := cmd.Run(); err != nil {
			continue // Skip files that can't be diffed
		}
		
		diffs = append(diffs, Diff{
			FilePath: file,
			Content:  out.String(),
		})
	}
	
	return diffs, nil
}

// getAllDiffs gets git diff for all changed files
func getAllDiffs() ([]Diff, error) {
	cmd := exec.Command("git", "diff", "--unified=0", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	
	return []Diff{
		{
			FilePath: "multiple_files",
			Content:  out.String(),
		},
	}, nil
}