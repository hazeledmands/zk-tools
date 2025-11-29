package zettel

import (
	"os"
	"path/filepath"
)

// GetZettelDir returns the zettelkasten directory from environment or default
func GetZettelDir() string {
	zettelDir := os.Getenv("ZETTEL_DIR")
	if zettelDir == "" {
		zettelDir = filepath.Join(os.Getenv("HOME"), "Projects", "Zettelkasten")
	}
	return zettelDir
}

// GetZettelDirFromArgs returns the zettelkasten directory from command-line args or environment
func GetZettelDirFromArgs(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	return GetZettelDir()
}

// GetMarkdownFiles returns all markdown files in the zettelkasten directory
func GetMarkdownFiles(zettelDir string) ([]string, error) {
	pattern := filepath.Join(zettelDir, "*.md")
	return filepath.Glob(pattern)
}

// ShouldSkipFile returns true if the file should be skipped in processing
func ShouldSkipFile(filename string, skipFiles ...string) bool {
	base := filepath.Base(filename)
	defaultSkip := []string{"00-index.md", "01-todo.md"}

	// Combine default skip list with custom skip files
	allSkip := append(defaultSkip, skipFiles...)

	for _, skip := range allSkip {
		if base == skip {
			return true
		}
	}
	return false
}
