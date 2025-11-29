package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexGeneration(t *testing.T) {
	// Create a temporary directory for test notes
	tmpDir := t.TempDir()

	// Create test notes
	notes := map[string]string{
		"note1.md": `# Programming Concepts
- #golang
- #patterns

Programming is fundamental to software development.

## Design Patterns
Factory and singleton patterns.
`,
		"note2.md": `# Data Structures
- #algorithms
- #complexity

Understanding data structures is crucial.

## Arrays
Fixed size collections.

## Linked Lists
Dynamic size collections.
`,
		"note3.md": `# Functional Programming
- #golang
- #functional

FP emphasizes immutability.
`,
	}

	for filename, content := range notes {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to use temp directory
	os.Args = []string{"zettelkasten-index", tmpDir}

	// Run main
	main()

	// Verify index file was created
	indexPath := filepath.Join(tmpDir, "00-index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatalf("Index file was not created")
	}

	// Read and verify index content
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}

	indexContent := string(content)

	// Verify index has the expected sections
	if !strings.Contains(indexContent, "# Zettelkasten Index") {
		t.Error("Index missing main heading")
	}

	if !strings.Contains(indexContent, "## Keywords") {
		t.Error("Index missing Keywords section")
	}

	if !strings.Contains(indexContent, "## All Headings") {
		t.Error("Index missing All Headings section")
	}

	// Verify keywords are present
	expectedKeywords := []string{"#golang", "#patterns", "#algorithms", "#complexity", "#functional"}
	for _, keyword := range expectedKeywords {
		if !strings.Contains(indexContent, keyword) {
			t.Errorf("Index missing keyword: %s", keyword)
		}
	}

	// Verify headings are present
	expectedHeadings := []string{
		"[[note1#Programming Concepts]]",
		"[[note2#Data Structures]]",
		"[[note3#Functional Programming]]",
		"[[note2#Arrays]]",
		"[[note2#Linked Lists]]",
	}

	for _, heading := range expectedHeadings {
		if !strings.Contains(indexContent, heading) {
			t.Errorf("Index missing heading: %s", heading)
		}
	}

	// Verify subheadings appear with proper indentation
	// Design Patterns should appear as a link (it's an H2 under "Programming Concepts")
	linkText := "[[note1#Design Patterns]]"
	if !strings.Contains(indexContent, linkText) {
		t.Error("Index should contain subheading link: [[note1#Design Patterns]]")
	}
}

func TestIndexGeneration_EmptyDirectory(t *testing.T) {
	t.Skip("Skipping test that calls os.Exit - main() calls os.Exit(0) for empty directories")
	// This test would fail because main() calls os.Exit(0) when no files found
	// To make this testable, we'd need to refactor main() to not call os.Exit during tests
}

func TestIndexGeneration_SkipsIndexFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test note and an existing index
	testNote := filepath.Join(tmpDir, "test.md")
	os.WriteFile(testNote, []byte("# Test Note\n"), 0644)

	existingIndex := filepath.Join(tmpDir, "00-index.md")
	os.WriteFile(existingIndex, []byte("# Old Index\nOld content\n"), 0644)

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-index", tmpDir}

	// Run main
	main()

	// Read new index
	content, err := os.ReadFile(existingIndex)
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}

	indexContent := string(content)

	// Should NOT contain a self-reference to 00-index
	if strings.Contains(indexContent, "[[00-index") {
		t.Error("Index should not reference itself")
	}

	// Should contain the test note
	if !strings.Contains(indexContent, "[[test#Test Note]]") {
		t.Error("Index should contain test note")
	}
}

func TestIndexGeneration_HandlesNoKeywords(t *testing.T) {
	tmpDir := t.TempDir()

	// Create note without keywords
	noteWithoutKeywords := filepath.Join(tmpDir, "plain.md")
	os.WriteFile(noteWithoutKeywords, []byte("# Plain Note\n\nNo keywords here.\n"), 0644)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"zettelkasten-index", tmpDir}

	main()

	// Verify index was created
	indexPath := filepath.Join(tmpDir, "00-index.md")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}

	indexContent := string(content)

	// Should still have the heading
	if !strings.Contains(indexContent, "[[plain#Plain Note]]") {
		t.Error("Index should contain heading even without keywords")
	}
}
