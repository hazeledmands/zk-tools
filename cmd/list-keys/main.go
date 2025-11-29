package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazel/zk-tools/internal/markdown"
	"github.com/hazel/zk-tools/internal/zettel"
)

var (
	dryRun   = flag.Bool("dry-run", false, "Show what would be changed without modifying files")
	backup   = flag.Bool("backup", true, "Create backup before modifying files")
	noBackup = flag.Bool("no-backup", false, "Skip backup creation (overrides --backup)")
	yes      = flag.Bool("yes", false, "Skip confirmation prompt")
	restore  = flag.Bool("restore", false, "Restore from most recent backup")
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

// analyzeChanges returns a description of what would change for dry-run
func analyzeChanges(filePath string) (string, error) {
	keywords, err := markdown.FrontMatterKeywords(filePath)
	if err != nil {
		return "", err
	}

	if len(keywords) == 0 {
		return "No keywords to move", nil
	}

	return fmt.Sprintf("Add keywords: %s", strings.Join(keywords, ", ")), nil
}

func main() {
	flag.Parse()

	// Get directory from remaining args or env
	var zettelDir string
	if flag.NArg() > 0 {
		zettelDir = flag.Arg(0)
	} else {
		zettelDir = zettel.GetZettelDir()
	}

	// Handle restore mode
	if *restore {
		backupDir, err := zettel.FindLatestBackup(zettelDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		timestamp, _ := zettel.GetBackupTimestamp(backupDir)
		fmt.Printf("Found backup from %s\n", timestamp.Format("2006-01-02 15:04:05"))

		if !*yes {
			if !zettel.AskForConfirmation("Restore files from this backup?") {
				fmt.Println("Restore cancelled")
				os.Exit(0)
			}
		}

		err = zettel.RestoreFromBackup(backupDir, zettelDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error restoring: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Restored all files from backup")
		return
	}

	files, err := zettel.GetMarkdownFiles(zettelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding markdown files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No markdown files found in %s\n", zettelDir)
		os.Exit(0)
	}

	// Dry-run mode
	if *dryRun {
		fmt.Println("DRY RUN: Showing what would be changed")
		fmt.Println()

		changeCount := 0
		for _, file := range files {
			changes, err := analyzeChanges(file)
			if err != nil {
				continue // Skip files without valid front-matter
			}

			if changes != "No keywords to move" {
				fmt.Printf("Would modify: %s\n", filepath.Base(file))
				fmt.Printf("  - %s\n", changes)
				fmt.Printf("  - Remove front-matter\n\n")
				changeCount++
			}
		}

		if changeCount == 0 {
			fmt.Println("No files would be modified")
		} else {
			fmt.Printf("Would modify %d files\n", changeCount)
		}
		return
	}

	// Determine if backup should be created
	shouldBackup := *backup && !*noBackup

	// Create backup before modification
	if shouldBackup {
		fmt.Println("Creating backup...")
		backupDir, err := zettel.CreateBackup(zettelDir, zettel.BackupOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created backup at %s\n\n", backupDir)
	}

	// Count files that will be modified
	modifyCount := 0
	for _, file := range files {
		_, err := markdown.FrontMatterKeywords(file)
		if err == nil {
			modifyCount++
		}
	}

	// Ask for confirmation
	if !*yes && modifyCount > 0 {
		prompt := fmt.Sprintf("This will modify %d files. Continue?", modifyCount)
		if !zettel.AskForConfirmation(prompt) {
			fmt.Println("Operation cancelled")
			os.Exit(0)
		}
	}

	processedCount := 0
	errorCount := 0

	fmt.Println("Moving keywords from front-matter to below first heading...")
	fmt.Println()

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
