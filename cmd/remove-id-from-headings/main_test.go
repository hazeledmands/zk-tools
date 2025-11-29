package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveIDFromHeading(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "H1 with ID",
			input: `# 202401011200: My Note Title

Content here.
`,
			expected: `# My Note Title

Content here.
`,
		},
		{
			name: "H2 with ID",
			input: `## 202401011200: Section Title

Section content.
`,
			expected: `## Section Title

Section content.
`,
		},
		{
			name: "Multiple headings with IDs",
			input: `# 202401011200: Main Title

Content.

## 202401011201: Subsection

More content.

### 202401011202: Sub-subsection

Even more content.
`,
			expected: `# Main Title

Content.

## Subsection

More content.

### Sub-subsection

Even more content.
`,
		},
		{
			name: "Headings without IDs (unchanged)",
			input: `# Regular Heading

Content.

## Another Heading

More content.
`,
			expected: `# Regular Heading

Content.

## Another Heading

More content.
`,
		},
		{
			name: "Mixed: some with IDs, some without",
			input: `# 202401011200: With ID

Content.

## Regular Heading

More content.

### 202401011201: Also with ID

Final content.
`,
			expected: `# With ID

Content.

## Regular Heading

More content.

### Also with ID

Final content.
`,
		},
		{
			name: "ID with different formats",
			input: `# 123456: Short ID

## 20240101: Date-like ID

### 999999999999: Long ID

Content.
`,
			expected: `# Short ID

## Date-like ID

### Long ID

Content.
`,
		},
		{
			name: "Preserve content with numbers",
			input: `# 202401011200: Title with 123 numbers

Content with 456 more numbers.

## Regular section with 789

More content.
`,
			expected: `# Title with 123 numbers

Content with 456 more numbers.

## Regular section with 789

More content.
`,
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

			err = removeIDFromHeading(testFile)
			if err != nil {
				t.Fatalf("Error processing file: %v", err)
			}

			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			if string(content) != tt.expected {
				t.Errorf("Output doesn't match expected.\nGot:\n%s\n\nExpected:\n%s", string(content), tt.expected)
			}
		})
	}
}

func TestRemoveIDFromHeading_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"note1.md": `# 202401011200: First Note

This is the first note with an ID.

## 202401011201: Subsection

Subsection content.
`,
		"note2.md": `# Regular Note

This note has no IDs to remove.
`,
		"note3.md": `# 202401011300: Another Note

## Regular Subsection

### 202401011301: Nested Section

Mixed content.
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

	// Use --yes to skip confirmation and --no-backup for cleaner tests
	os.Args = []string{"remove-id-from-headings", "--yes", "--no-backup", tmpDir}

	// Run main
	main()

	// Verify note1.md was modified
	content1, err := os.ReadFile(filepath.Join(tmpDir, "note1.md"))
	if err != nil {
		t.Fatalf("Failed to read note1.md: %v", err)
	}

	note1Content := string(content1)

	if strings.Contains(note1Content, "202401011200:") {
		t.Error("note1.md should not contain ID '202401011200:'")
	}

	if !strings.Contains(note1Content, "# First Note") {
		t.Error("note1.md should contain '# First Note'")
	}

	if !strings.Contains(note1Content, "## Subsection") {
		t.Error("note1.md should contain '## Subsection'")
	}

	// Verify note2.md is unchanged (no IDs to remove)
	content2, err := os.ReadFile(filepath.Join(tmpDir, "note2.md"))
	if err != nil {
		t.Fatalf("Failed to read note2.md: %v", err)
	}

	if string(content2) != files["note2.md"] {
		t.Error("note2.md should be unchanged")
	}

	// Verify note3.md had IDs removed
	content3, err := os.ReadFile(filepath.Join(tmpDir, "note3.md"))
	if err != nil {
		t.Fatalf("Failed to read note3.md: %v", err)
	}

	note3Content := string(content3)

	if strings.Contains(note3Content, "202401011300:") || strings.Contains(note3Content, "202401011301:") {
		t.Error("note3.md should not contain any IDs")
	}

	if !strings.Contains(note3Content, "# Another Note") {
		t.Error("note3.md should contain '# Another Note'")
	}

	if !strings.Contains(note3Content, "### Nested Section") {
		t.Error("note3.md should contain '### Nested Section'")
	}
}

func TestRemoveIDFromHeading_PreservesFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	input := `# 202401011200: Title

Paragraph with **bold** and *italic* text.

## 202401011201: Code Section

` + "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```" + `

### 202401011202: List Section

- Item 1
- Item 2
  - Nested item

Regular text continues here.

Link to [[another-note]] should be preserved.

> Blockquote text should be preserved.

End of document.
`

	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte(input), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = removeIDFromHeading(testFile)
	if err != nil {
		t.Fatalf("Error processing file: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	output := string(content)

	// Verify IDs are removed
	if strings.Contains(output, "202401011200:") || strings.Contains(output, "202401011201:") || strings.Contains(output, "202401011202:") {
		t.Error("IDs should be removed from headings")
	}

	// Verify formatting is preserved
	mustContain := []string{
		"# Title",
		"**bold**",
		"*italic*",
		"## Code Section",
		"func main() {",
		"### List Section",
		"- Item 1",
		"- Item 2",
		"  - Nested item",
		"[[another-note]]",
		"> Blockquote text",
		"End of document.",
	}

	for _, text := range mustContain {
		if !strings.Contains(output, text) {
			t.Errorf("Output missing expected text: %q", text)
		}
	}
}

func TestRemoveIDFromHeading_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty file",
			input:    "",
			expected: "",
		},
		{
			name: "Only heading with ID",
			input: `# 123456: Title
`,
			expected: `# Title
`,
		},
		{
			name: "Heading with ID at end of file",
			input: `# 123456: Title`,
			expected: `# Title
`,
		},
		{
			name: "Heading with colon in title (no ID)",
			input: `# My Note: A Subtitle

Content.
`,
			expected: `# My Note: A Subtitle

Content.
`,
		},
		{
			name: "Multiple colons",
			input: `# 123456: Title: With: Colons

Content.
`,
			expected: `# Title: With: Colons

Content.
`,
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

			err = removeIDFromHeading(testFile)
			if err != nil {
				t.Fatalf("Error processing file: %v", err)
			}

			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if string(content) != tt.expected {
				t.Errorf("Output doesn't match.\nGot:\n%q\n\nExpected:\n%q", string(content), tt.expected)
			}
		})
	}
}
