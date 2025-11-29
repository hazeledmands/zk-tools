package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountLinksInLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "No links",
			input:    "This is just plain text",
			expected: 0,
		},
		{
			name:     "Single link",
			input:    "See [[other-note]] for details",
			expected: 1,
		},
		{
			name:     "Multiple links",
			input:    "Check [[note1]] and [[note2]] and [[note3]]",
			expected: 3,
		},
		{
			name:     "Link with heading",
			input:    "See [[note#heading]] for more",
			expected: 1,
		},
		{
			name:     "Malformed link - no closing",
			input:    "This has [[unclosed link",
			expected: 0,
		},
		{
			name:     "Empty link",
			input:    "This has [[]] empty link",
			expected: 1,
		},
		{
			name:     "Multiple links on same line",
			input:    "[[first]] text [[second]] more [[third]]",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountLinksInLine(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d links, got %d", tt.expected, result)
			}
		})
	}
}

func TestExtractLinksFromLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "No links",
			input:    "Plain text without links",
			expected: []string{},
		},
		{
			name:     "Single link",
			input:    "See [[other-note]] for details",
			expected: []string{"other-note"},
		},
		{
			name:     "Link with heading - should extract only file ID",
			input:    "See [[note#heading]] for details",
			expected: []string{"note"},
		},
		{
			name:     "Multiple links",
			input:    "Check [[note1]] and [[note2#section]] and [[note3]]",
			expected: []string{"note1", "note2", "note3"},
		},
		{
			name:     "Empty link - should be skipped",
			input:    "This has [[]] empty",
			expected: []string{},
		},
		{
			name:     "Link with alias",
			input:    "See [[actual-note|Display Text]]",
			expected: []string{"actual-note|Display Text"},
		},
		{
			name:     "Malformed link - no closing",
			input:    "[[unclosed",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractLinksFromLine(tt.input)
			// Handle nil vs empty slice comparison
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func TestCountBacklinks(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test files with links
	file1 := filepath.Join(tmpDir, "note1.md")
	file2 := filepath.Join(tmpDir, "note2.md")
	file3 := filepath.Join(tmpDir, "note3.md")

	// note1 links to note2 twice and note3 once
	content1 := `# Note 1
See [[note2]] for more info.
Also check [[note2#section]] again.
And [[note3]] is useful too.
`

	// note2 links to note3 once
	content2 := `# Note 2
Refer to [[note3]] for details.
`

	// note3 has no links
	content3 := `# Note 3
This note has no outgoing links.
`

	err := os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(file3, []byte(content3), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Count backlinks
	files := []string{file1, file2, file3}
	backlinks, err := CountBacklinks(files)
	if err != nil {
		t.Fatalf("Failed to count backlinks: %v", err)
	}

	// Verify backlink counts
	// note1 has 0 backlinks (nobody links to it)
	if backlinks["note1"] != 0 {
		t.Errorf("Expected note1 to have 0 backlinks, got %d", backlinks["note1"])
	}

	// note2 has 2 backlinks (note1 links to it twice)
	if backlinks["note2"] != 2 {
		t.Errorf("Expected note2 to have 2 backlinks, got %d", backlinks["note2"])
	}

	// note3 has 2 backlinks (note1 and note2 each link to it once)
	if backlinks["note3"] != 2 {
		t.Errorf("Expected note3 to have 2 backlinks, got %d", backlinks["note3"])
	}
}

func TestExtractNoteInfo(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "202401011200.md")
	content := `---
title: Test Note
keywords:
  - testing
  - automation
---

# Test Note
- #todo

This is a test note with some content.
It has multiple lines to increase word count.

## Section 1
- #todo

More content here with a link to [[other-note]].

## Section 2
Another section without todos.
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract note info
	noteInfo, err := ExtractNoteInfo(testFile)
	if err != nil {
		t.Fatalf("Failed to extract note info: %v", err)
	}

	// Verify file ID
	if noteInfo.FileID != "202401011200" {
		t.Errorf("Expected FileID '202401011200', got %q", noteInfo.FileID)
	}

	// Verify title
	if noteInfo.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got %q", noteInfo.Title)
	}

	// Verify TODO headings (should have 2: "Test Note" and "Section 1")
	if len(noteInfo.TODOHeadings) != 2 {
		t.Errorf("Expected 2 TODO headings, got %d", len(noteInfo.TODOHeadings))
	}

	// Verify outgoing links
	if noteInfo.OutgoingLinkCount != 1 {
		t.Errorf("Expected 1 outgoing link, got %d", noteInfo.OutgoingLinkCount)
	}

	// Verify word count is greater than 0
	if noteInfo.WordCount <= 0 {
		t.Errorf("Expected positive word count, got %d", noteInfo.WordCount)
	}
}

func TestExtractNoteInfo_NoTitle(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test-note.md")
	content := `Some content without a heading.
Just plain text.
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	noteInfo, err := ExtractNoteInfo(testFile)
	if err != nil {
		t.Fatalf("Failed to extract note info: %v", err)
	}

	// Should use filename as title
	if noteInfo.Title != "test-note" {
		t.Errorf("Expected title to be filename 'test-note', got %q", noteInfo.Title)
	}
}

func TestExtractNoteInfo_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "empty.md")
	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	noteInfo, err := ExtractNoteInfo(testFile)
	if err != nil {
		t.Fatalf("Failed to extract note info: %v", err)
	}

	// Verify defaults for empty file
	if noteInfo.WordCount != 0 {
		t.Errorf("Expected 0 word count, got %d", noteInfo.WordCount)
	}

	if noteInfo.OutgoingLinkCount != 0 {
		t.Errorf("Expected 0 outgoing links, got %d", noteInfo.OutgoingLinkCount)
	}

	if len(noteInfo.TODOHeadings) != 0 {
		t.Errorf("Expected 0 TODO headings, got %d", len(noteInfo.TODOHeadings))
	}
}
