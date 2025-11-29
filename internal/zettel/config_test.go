package zettel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetZettelDir(t *testing.T) {
	// Save original environment
	originalZettelDir := os.Getenv("ZETTEL_DIR")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("ZETTEL_DIR", originalZettelDir)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name            string
		zettelDir       string
		home            string
		expectedContain string
	}{
		{
			name:            "ZETTEL_DIR set",
			zettelDir:       "/custom/zettel/path",
			home:            "/home/user",
			expectedContain: "/custom/zettel/path",
		},
		{
			name:            "ZETTEL_DIR not set, use default",
			zettelDir:       "",
			home:            "/home/testuser",
			expectedContain: "/home/testuser/Projects/Zettelkasten",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ZETTEL_DIR", tt.zettelDir)
			os.Setenv("HOME", tt.home)

			result := GetZettelDir()

			if result != tt.expectedContain {
				t.Errorf("Expected %q, got %q", tt.expectedContain, result)
			}
		})
	}
}

func TestGetZettelDirFromArgs(t *testing.T) {
	// Save original environment
	originalZettelDir := os.Getenv("ZETTEL_DIR")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("ZETTEL_DIR", originalZettelDir)
		os.Setenv("HOME", originalHome)
	}()

	os.Setenv("ZETTEL_DIR", "/env/zettel")
	os.Setenv("HOME", "/home/user")

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "No arguments - use env",
			args:     []string{"program"},
			expected: "/env/zettel",
		},
		{
			name:     "One argument - use env",
			args:     []string{"program"},
			expected: "/env/zettel",
		},
		{
			name:     "Two arguments - use second arg",
			args:     []string{"program", "/custom/path"},
			expected: "/custom/path",
		},
		{
			name:     "Multiple arguments - use second arg",
			args:     []string{"program", "/custom/path", "extra"},
			expected: "/custom/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetZettelDirFromArgs(tt.args)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	files := []string{
		"note1.md",
		"note2.md",
		"note3.md",
		"readme.txt",
		"index.html",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		err := os.WriteFile(path, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Get markdown files
	mdFiles, err := GetMarkdownFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get markdown files: %v", err)
	}

	// Should find exactly 3 .md files
	if len(mdFiles) != 3 {
		t.Errorf("Expected 3 markdown files, got %d", len(mdFiles))
	}

	// Verify all returned files are .md files
	for _, file := range mdFiles {
		if filepath.Ext(file) != ".md" {
			t.Errorf("Expected .md file, got %q", file)
		}
	}
}

func TestGetMarkdownFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	mdFiles, err := GetMarkdownFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get markdown files: %v", err)
	}

	if len(mdFiles) != 0 {
		t.Errorf("Expected 0 markdown files in empty directory, got %d", len(mdFiles))
	}
}

func TestGetMarkdownFiles_NonexistentDirectory(t *testing.T) {
	mdFiles, err := GetMarkdownFiles("/nonexistent/directory")
	// Should not error, just return empty list
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mdFiles) != 0 {
		t.Errorf("Expected 0 files for nonexistent directory, got %d", len(mdFiles))
	}
}

func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		skipFiles []string
		expected  bool
	}{
		{
			name:      "Skip default: 00-index.md",
			filename:  "/path/to/00-index.md",
			skipFiles: []string{},
			expected:  true,
		},
		{
			name:      "Skip default: 01-todo.md",
			filename:  "/path/to/01-todo.md",
			skipFiles: []string{},
			expected:  true,
		},
		{
			name:      "Don't skip regular file",
			filename:  "/path/to/regular-note.md",
			skipFiles: []string{},
			expected:  false,
		},
		{
			name:      "Skip custom file",
			filename:  "/path/to/draft.md",
			skipFiles: []string{"draft.md"},
			expected:  true,
		},
		{
			name:      "Skip multiple custom files",
			filename:  "/path/to/temp.md",
			skipFiles: []string{"draft.md", "temp.md", "archive.md"},
			expected:  true,
		},
		{
			name:      "Don't skip when not in list",
			filename:  "/path/to/note.md",
			skipFiles: []string{"draft.md", "temp.md"},
			expected:  false,
		},
		{
			name:      "Path includes similar name but different file",
			filename:  "/00-index/regular.md",
			skipFiles: []string{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSkipFile(tt.filename, tt.skipFiles...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldSkipFile_WithBothDefaultAndCustom(t *testing.T) {
	// Test that both default and custom skip files work together
	tests := []struct {
		filename  string
		skipFiles []string
		expected  bool
	}{
		{"00-index.md", []string{"custom.md"}, true},  // Default skip
		{"01-todo.md", []string{"custom.md"}, true},   // Default skip
		{"custom.md", []string{"custom.md"}, true},    // Custom skip
		{"regular.md", []string{"custom.md"}, false},  // Don't skip
	}

	for _, tt := range tests {
		result := ShouldSkipFile(tt.filename, tt.skipFiles...)
		if result != tt.expected {
			t.Errorf("File %q with custom %v: expected %v, got %v",
				tt.filename, tt.skipFiles, tt.expected, result)
		}
	}
}
