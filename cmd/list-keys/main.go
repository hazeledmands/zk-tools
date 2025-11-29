package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazel/zk-tools/internal/markdown"
	"github.com/hazel/zk-tools/internal/zettel"
)

// moveKeywordsBelowHeading moves keywords from front-matter to below the first heading
func moveKeywordsBelowHeading(filePath string) error {
	// Extract keywords from front-matter
	keywords, err := markdown.FrontMatterKeywords(filePath)
	if err != nil {
		return err
	}

	// Read content after front-matter
	contentLines, err := markdown.ReadContentAfterFrontMatter(filePath)
	if err != nil {
		return err
	}

	// Find the first heading and insert keywords after it
	var newContentLines []string
	firstHeadingFound := false

	for _, line := range contentLines {
		newContentLines = append(newContentLines, line)

		// Check if this is the first heading
		if !firstHeadingFound && strings.HasPrefix(strings.TrimSpace(line), "#") {
			firstHeadingFound = true

			// Always add a blank line after heading
			newContentLines = append(newContentLines, "")

			// Add keywords as bulleted list
			for _, keyword := range keywords {
				newContentLines = append(newContentLines, "- "+keyword)
			}
		}
	}

	// Build the new file content (no front-matter)
	var newContent strings.Builder
	newContent.WriteString(strings.Join(newContentLines, "\n"))
	if len(newContentLines) > 0 {
		newContent.WriteString("\n")
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func main() {
	zettelDir := zettel.GetZettelDirFromArgs(os.Args)

	files, err := zettel.GetMarkdownFiles(zettelDir)
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

	fmt.Printf("Moving keywords from front-matter to below first heading...\n\n")

	for _, file := range files {
		if err := moveKeywordsBelowHeading(file); err != nil {
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
