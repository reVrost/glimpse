package git

import (
	"bytes"
	"os/exec"
	"strings"
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
	
	// Get diff for specific files (including staged)
	for _, file := range files {
		// First check unstaged changes
		cmd := exec.Command("git", "diff", "--unified=3", "HEAD", "--", file)
		var out bytes.Buffer
		cmd.Stdout = &out
		
		var content string
		if err := cmd.Run(); err == nil {
			content = out.String()
		}
		
		// Then check staged changes
		cmd2 := exec.Command("git", "diff", "--unified=3", "--cached", "HEAD", "--", file)
		var stagedOut bytes.Buffer
		cmd2.Stdout = &stagedOut
		
		if err := cmd2.Run(); err == nil {
			if content != "" {
				content += "\n" + stagedOut.String()
			} else {
				content = stagedOut.String()
			}
		}
		
		// Only add if there's actual content
		if content != "" {
			diffs = append(diffs, Diff{
				FilePath: file,
				Content:  content,
			})
		}
	}
	
	return diffs, nil
}

// getAllDiffs gets git diff for all changed files
func getAllDiffs() ([]Diff, error) {
	// Get diff for both staged and unstaged changes
	cmd := exec.Command("git", "diff", "--unified=3", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	
	// Also get staged changes if any
	cmd2 := exec.Command("git", "diff", "--unified=3", "--cached", "HEAD")
	var stagedOut bytes.Buffer
	cmd2.Stdout = &stagedOut
	
	if err := cmd2.Run(); err == nil {
		// Combine both diffs
		stagedContent := stagedOut.String()
		if stagedContent != "" {
			return []Diff{
				{
					FilePath: "all_changes",
					Content:  out.String() + "\n" + stagedContent,
				},
			}, nil
		}
	}
	return []Diff{
		{
			FilePath: "all_changes",
			Content:  out.String(),
		},
	}, nil
}

// GetChangedFiles returns a list of all changed files (staged and unstaged)
func GetChangedFiles() ([]string, error) {
	var changedFiles []string
	
	// Get unstaged changes
	cmd1 := exec.Command("git", "diff", "--name-only", "HEAD")
	var out1 bytes.Buffer
	cmd1.Stdout = &out1
	
	if err := cmd1.Run(); err != nil {
		return nil, err
	}
	
	// Get staged changes
	cmd2 := exec.Command("git", "diff", "--name-only", "--cached", "HEAD")
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	
	if err := cmd2.Run(); err != nil {
		return nil, err
	}
	
	// Parse and deduplicate file lists
	fileMap := make(map[string]bool)
	
	// Parse unstaged files
	if out1.String() != "" {
		files := strings.Split(out1.String(), "\n")
		for _, file := range files {
			if file != "" {
				fileMap[file] = true
			}
		}
	}
	
	// Parse staged files
	if out2.String() != "" {
		files := strings.Split(out2.String(), "\n")
		for _, file := range files {
			if file != "" {
				fileMap[file] = true
			}
		}
	}
	
	// Convert back to slice
	for file := range fileMap {
		changedFiles = append(changedFiles, file)
	}
	
	return changedFiles, nil
}