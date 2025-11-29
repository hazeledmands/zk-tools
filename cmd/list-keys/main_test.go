package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveKeywordsBelowHeading(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLines []string
		shouldError   bool
	}{
		{
			name: "Valid front-matter with keywords",
			input: `---
title: Test Note
keywords:
  - golang
  - testing
  - automation
---

# My Note

Content here.
`,
			expectedLines: []string{
				"# My Note",
				"",
				"- golang",
				"- testing",
				"- automation",
				"",
				"Content here.",
			},
			shouldError: false,
		},
		{
			name: "Front-matter without keywords",
			input: `---
title: Test Note
author: John
---

# My Note

Content here.
`,
			expectedLines: []string{
				"# My Note",
				"",
				"Content here.",
			},
			shouldError: false,
		},
		{
			name: "Multiple headings - should insert after first only",
			input: `---
keywords:
  - test
---

# First Heading

Some content.

## Second Heading

More content.
`,
			expectedLines: []string{
				"# First Heading",
				"",
				"- test",
				"",
				"Some content.",
				"",
				"## Second Heading",
			},
			shouldError: false,
		},
		{
			name: "No front-matter",
			input: `# Just a heading

Content without front-matter.
`,
			expectedLines: nil,
			shouldError:   true,
		},
		{
			name: "No heading in content",
			input: `---
keywords:
  - test
---

Just content without any heading.
`,
			expectedLines: []string{
				"",
				"Just content without any heading.",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.md")

			err := os.WriteFile(testFile, []byte(tt.input), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err = moveKeywordsBelowHeading(testFile)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Read the modified file
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")

			// Verify expected lines are present
			for _, expectedLine := range tt.expectedLines {
				found := false
				for _, line := range lines {
					if line == expectedLine {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected line %q not found in output:\n%s", expectedLine, string(content))
				}
			}

			// Verify front-matter is removed
			if strings.Contains(string(content), "---") {
				t.Error("Front-matter delimiters should be removed")
			}
		})
	}
}

func TestListKeysIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files (only files with valid front-matter to avoid os.Exit)
	files := map[string]string{
		"note1.md": `---
title: First Note
keywords:
  - golang
  - testing
---

# First Note

Content here.
`,
		"note2.md": `---
keywords:
  - python
---

# Second Note

More content.
`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"list-keys", tmpDir}

	// Run main - should succeed with these files
	main()

	// Verify note1.md was modified correctly
	content1, err := os.ReadFile(filepath.Join(tmpDir, "note1.md"))
	if err != nil {
		t.Fatalf("Failed to read note1.md: %v", err)
	}

	note1Content := string(content1)

	if strings.Contains(note1Content, "---") {
		t.Error("note1.md should not contain front-matter delimiters")
	}

	if !strings.Contains(note1Content, "- golang") {
		t.Error("note1.md should contain '- golang'")
	}

	if !strings.Contains(note1Content, "- testing") {
		t.Error("note1.md should contain '- testing'")
	}

	// Verify note2.md was modified
	content2, err := os.ReadFile(filepath.Join(tmpDir, "note2.md"))
	if err != nil {
		t.Fatalf("Failed to read note2.md: %v", err)
	}

	note2Content := string(content2)

	if strings.Contains(note2Content, "---") {
		t.Error("note2.md should not contain front-matter delimiters")
	}

	if !strings.Contains(note2Content, "- python") {
		t.Error("note2.md should contain '- python'")
	}
}

func TestListKeys_PreservesContent(t *testing.T) {
	tmpDir := t.TempDir()

	input := `---
keywords:
  - test
  - example
---

# Important Note

This is important content that must be preserved exactly.
Including special characters: & < > " '
And [[links]] to other notes.

## Subsection

- Bullet point 1
- Bullet point 2

Code block:
` + "```go\nfunc main() {\n}\n```" + `

End of note.
`

	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte(input), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = moveKeywordsBelowHeading(testFile)
	if err != nil {
		t.Fatalf("Error processing file: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	output := string(content)

	// Verify all important content is preserved
	mustContain := []string{
		"# Important Note",
		"This is important content that must be preserved exactly.",
		"Including special characters: & < > \" '",
		"And [[links]] to other notes.",
		"## Subsection",
		"- Bullet point 1",
		"- Bullet point 2",
		"func main() {",
		"End of note.",
	}

	for _, text := range mustContain {
		if !strings.Contains(output, text) {
			t.Errorf("Output missing expected text: %q", text)
		}
	}

	// Verify keywords were added
	if !strings.Contains(output, "- test") {
		t.Error("Keywords not added correctly")
	}

	if !strings.Contains(output, "- example") {
		t.Error("Keywords not added correctly")
	}
}
