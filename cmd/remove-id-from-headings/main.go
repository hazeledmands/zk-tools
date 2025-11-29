package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hazel/zk-tools/internal/zettel"
)

// removeIDFromHeading removes the ID prefix from headings in the format "# ID: Title"
func removeIDFromHeading(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Pattern to match headings like "# 202003092000: Wikipedia Streams"
	headingPattern := regexp.MustCompile(`^(#+)\s+\d+:\s+(.+)$`)

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line matches the pattern
		if matches := headingPattern.FindStringSubmatch(line); matches != nil {
			// matches[1] is the heading markers (# or ## or ###, etc.)
			// matches[2] is the title
			line = matches[1] + " " + matches[2]
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Write the file back
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
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

	fmt.Printf("Removing IDs from headings...\n\n")

	for _, file := range files {
		if err := removeIDFromHeading(file); err != nil {
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
