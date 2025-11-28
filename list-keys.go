package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// rewriteFrontMatter rewrites a file's front-matter to keep only the keywords field
func rewriteFrontMatter(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var frontMatterLines []string
	var contentLines []string
	inFrontMatter := false
	frontMatterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			frontMatterCount++
			if frontMatterCount == 1 {
				inFrontMatter = true
				continue
			} else if frontMatterCount == 2 {
				inFrontMatter = false
				continue
			}
		}

		if inFrontMatter && frontMatterCount == 1 {
			frontMatterLines = append(frontMatterLines, line)
		} else if frontMatterCount >= 2 {
			contentLines = append(contentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if frontMatterCount < 2 {
		return fmt.Errorf("no valid YAML front-matter found")
	}

	// Extract keywords lines manually to preserve # characters
	var keywordLines []string
	inKeywords := false
	for _, line := range frontMatterLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "keywords:") {
			inKeywords = true
			keywordLines = append(keywordLines, line)
		} else if inKeywords {
			// Check if this is still part of the keywords list (starts with -)
			if strings.HasPrefix(trimmed, "-") {
				keywordLines = append(keywordLines, line)
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "-") {
				// We've reached a new key, stop collecting keywords
				break
			}
		}
	}

	// Build the new file content
	var newContent strings.Builder
	newContent.WriteString("---\n")
	if len(keywordLines) > 0 {
		newContent.WriteString(strings.Join(keywordLines, "\n"))
		newContent.WriteString("\n")
	}
	newContent.WriteString("---\n")
	newContent.WriteString(strings.Join(contentLines, "\n"))
	if len(contentLines) > 0 {
		newContent.WriteString("\n")
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func main() {
	zettelDir := os.Getenv("ZETTEL_DIR")
	if zettelDir == "" {
		zettelDir = filepath.Join(os.Getenv("HOME"), "Projects", "Zettelkasten")
	}

	if len(os.Args) > 1 {
		zettelDir = os.Args[1]
	}

	pattern := filepath.Join(zettelDir, "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding markdown files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No markdown files found in %s\n", zettelDir)
		os.Exit(0)
	}

	processedCount := 0
	errorCount := 0

	fmt.Printf("Rewriting front-matter to keep only 'keywords' field...\n\n")

	for _, file := range files {
		if err := rewriteFrontMatter(file); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", filepath.Base(file), err)
			errorCount++
		} else {
			fmt.Printf("✓ %s\n", filepath.Base(file))
			processedCount++
		}
	}

	fmt.Printf("\nProcessed %d files in %s\n", processedCount, zettelDir)

	if errorCount > 0 {
		fmt.Printf("%d files had errors and were skipped\n", errorCount)
		os.Exit(1)
	}
}
