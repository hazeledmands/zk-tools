package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrontMatterKeywords(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expected    []string
		shouldError bool
	}{
		{
			name: "Valid front-matter with keywords",
			content: `---
title: Test Note
keywords:
  - golang
  - testing
  - automation
---

# Content
`,
			expected:    []string{"golang", "testing", "automation"},
			shouldError: false,
		},
		{
			name: "Front-matter with no keywords",
			content: `---
title: Test Note
author: John Doe
---

# Content
`,
			expected:    []string{},
			shouldError: false,
		},
		{
			name: "Keywords with extra spaces",
			content: `---
keywords:
  -   golang
  -  testing
  - automation
---

# Content
`,
			expected:    []string{"golang", "testing", "automation"},
			shouldError: false,
		},
		{
			name: "No front-matter",
			content: `# Just a heading
Content without front-matter.
`,
			expected:    nil,
			shouldError: true,
		},
		{
			name: "Incomplete front-matter (only one delimiter)",
			content: `---
title: Test
keywords:
  - testing
`,
			expected:    nil,
			shouldError: true,
		},
		{
			name: "Empty keywords list",
			content: `---
title: Test
keywords:
---

# Content
`,
			expected:    []string{},
			shouldError: false,
		},
		{
			name: "Keywords inline format (not list)",
			content: `---
title: Test
keywords: golang, testing
---

# Content
`,
			expected:    []string{},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := FrontMatterKeywords(testFile)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keywords, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Expected keyword %q at position %d, got %q", tt.expected[i], i, result[i])
				}
			}
		})
	}
}

func TestReadContentAfterFrontMatter(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Normal file with front-matter",
			content: `---
title: Test
---

# Heading
Content line 1
Content line 2
`,
			expected: []string{"", "# Heading", "Content line 1", "Content line 2"},
		},
		{
			name: "File without front-matter",
			content: `# Heading
Content
`,
			expected: []string{},
		},
		{
			name: "Empty content after front-matter",
			content: `---
title: Test
---
`,
			expected: []string{},
		},
		{
			name: "Front-matter with multiple fields",
			content: `---
title: Test
author: John
keywords:
  - test
  - markdown
date: 2024-01-01
---

First line
Second line
`,
			expected: []string{"", "First line", "Second line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := ReadContentAfterFrontMatter(testFile)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Got: %v", result)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestReadContentAfterFrontMatter_NonexistentFile(t *testing.T) {
	_, err := ReadContentAfterFrontMatter("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestFrontMatterKeywords_NonexistentFile(t *testing.T) {
	_, err := FrontMatterKeywords("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}
