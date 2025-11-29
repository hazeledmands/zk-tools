package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTODOGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test notes with different characteristics
	notes := map[string]string{
		"complete-note.md": `# Complete Note
- #reference

This is a complete note with over 100 words to ensure it doesn't get flagged as short.
It has substantial content and provides detailed information about the topic.
Let's add more words to make sure we exceed the 100 word threshold.
The quick brown fox jumps over the lazy dog multiple times in this sentence.
We need to add even more content here to ensure this note is considered complete.
Additional sentences are being added to reach the word count requirement.
This note should not appear in the TODO list at all because it's complete.
More filler text to ensure we have enough words for testing purposes.

See also: [[other-note]] for more information.
`,
		"short-note.md": `# Short Note
- #important

This note is too short.
`,
		"orphan-note.md": `# Orphan Note

This note has no links to or from other notes.
It has enough words to not be flagged as short, but it's isolated.
Let's add more content to ensure adequate word count here too.
`,
		"todo-note.md": `# Note with TODOs
- #todo

This section needs work.

## Research Topic
- #todo

Need to research this more thoroughly.

## Complete Section

This section is done and has content.
`,
		"linked-note.md": `# Linked Note

This note links to [[complete-note]] and is linked from complete-note.
It has adequate content and is well connected in the knowledge graph.
Therefore it should not appear in the TODO list at all because it meets all criteria.
Let's add more words to ensure this meets the minimum threshold of one hundred words.
Additional filler content is being added here to make sure we exceed the word count.
The quick brown fox jumps over the lazy dog repeatedly in multiple sentences here.
We need substantial content to avoid being flagged as a short note in the system.
This paragraph contains enough words to meet the one hundred word minimum requirement.
More content ensures that this note will not be considered unfinished or incomplete.
`,
	}

	for filename, content := range notes {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-todo", tmpDir}

	// Run main
	main()

	// Verify TODO file was created
	todoPath := filepath.Join(tmpDir, "01-todo.md")
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		t.Fatalf("TODO file was not created")
	}

	// Read and verify TODO content
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read TODO file: %v", err)
	}

	todoContent := string(content)

	// Verify TODO has expected sections
	if !strings.Contains(todoContent, "# TODO: Unfinished Notes") {
		t.Error("TODO missing main heading")
	}

	if !strings.Contains(todoContent, "## Orphan Notes") {
		t.Error("TODO missing Orphan Notes section")
	}

	if !strings.Contains(todoContent, "## Notes with #todo Tags") {
		t.Error("TODO missing Notes with #todo Tags section")
	}

	if !strings.Contains(todoContent, "## Short Notes") {
		t.Error("TODO missing Short Notes section")
	}

	// Verify orphan note is detected
	if !strings.Contains(todoContent, "[[orphan-note#Orphan Note]]") {
		t.Error("Orphan note should be listed")
	}

	// Verify short note is detected
	if !strings.Contains(todoContent, "[[short-note#Short Note]]") {
		t.Error("Short note should be listed")
	}

	// Verify TODO-tagged note is detected
	if !strings.Contains(todoContent, "[[todo-note#Note with TODOs]]") {
		t.Error("Note with #todo should be listed")
	}

	// Verify complete note is NOT listed
	if strings.Contains(todoContent, "[[complete-note") {
		t.Error("Complete note should not be listed in TODO")
	}

	// Verify linked note is NOT listed (has backlinks and adequate content)
	if strings.Contains(todoContent, "[[linked-note") {
		t.Error("Linked note should not be listed in TODO")
	}
}

func TestTODOGeneration_BacklinkCounting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes with links
	notes := map[string]string{
		"hub-note.md": `# Hub Note
Short but popular.
`,
		"note1.md": `# Note 1
Links to [[hub-note]] once.
`,
		"note2.md": `# Note 2
Links to [[hub-note]] and also [[hub-note#Section]].
`,
		"note3.md": `# Note 3
Another link to [[hub-note]].
`,
	}

	for filename, content := range notes {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-todo", tmpDir}
	main()

	todoPath := filepath.Join(tmpDir, "01-todo.md")
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read TODO file: %v", err)
	}

	todoContent := string(content)

	// Hub note should be listed (it's short)
	if !strings.Contains(todoContent, "[[hub-note#Hub Note]]") {
		t.Error("Hub note should be listed as short")
	}

	// Should show backlink count (4 backlinks: 1+2+1)
	if !strings.Contains(todoContent, "4 backlinks") {
		t.Error("Hub note should show 4 backlinks")
	}
}

func TestTODOGeneration_SkipsMetaFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes including meta files
	notes := map[string]string{
		"regular.md": `# Regular
Short note.
`,
		"00-index.md": `# Index
This is an index file.
`,
		"01-todo.md": `# TODO
Old TODO list.
`,
	}

	for filename, content := range notes {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-todo", tmpDir}
	main()

	todoPath := filepath.Join(tmpDir, "01-todo.md")
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read TODO file: %v", err)
	}

	todoContent := string(content)

	// Should include regular note
	if !strings.Contains(todoContent, "[[regular#Regular]]") {
		t.Error("Regular note should be listed")
	}

	// Should NOT reference meta files
	if strings.Contains(todoContent, "[[00-index") {
		t.Error("Index file should not be listed")
	}

	if strings.Contains(todoContent, "01-todo") && strings.Contains(todoContent, "Old TODO") {
		t.Error("Old TODO file should not be self-referenced")
	}
}

func TestTODOGeneration_MultipleTODOSections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a note with multiple TODO sections
	noteContent := `# Project Plan
- #todo

Introduction section needs work.

## Phase 1
- #todo

Research needed here.

## Phase 2

This phase is complete.

## Phase 3
- #todo

Implementation pending.
`

	path := filepath.Join(tmpDir, "project.md")
	err := os.WriteFile(path, []byte(noteContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-todo", tmpDir}
	main()

	todoPath := filepath.Join(tmpDir, "01-todo.md")
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read TODO file: %v", err)
	}

	todoContent := string(content)

	// Should show count of sections with #todo (3 sections)
	if !strings.Contains(todoContent, "3 sections with #todo") {
		t.Error("Should count 3 sections with #todo tags")
	}
}
