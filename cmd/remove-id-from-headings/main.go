package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hazel/zk-tools/internal/zettel"
)

var (
	dryRun   = flag.Bool("dry-run", false, "Show what would be changed without modifying files")
	backup   = flag.Bool("backup", true, "Create backup before modifying files")
	noBackup = flag.Bool("no-backup", false, "Skip backup creation (overrides --backup)")
	yes      = flag.Bool("yes", false, "Skip confirmation prompt")
	restore  = flag.Bool("restore", false, "Restore from most recent backup")
)

var headingPattern = regexp.MustCompile(`^(#+)\s+\d+:\s+(.+)$`)

// removeIDFromHeading removes the ID prefix from headings in the format "# ID: Title"
func removeIDFromHeading(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

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

// analyzeChanges returns a description of what would change for dry-run
func analyzeChanges(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var changes []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := headingPattern.FindStringSubmatch(line); matches != nil {
			// Found a heading with ID
			changes = append(changes, fmt.Sprintf("Line %d: %s → %s %s",
				lineNum, strings.TrimSpace(line), matches[1], matches[2]))
		}
	}

	return changes, scanner.Err()
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

		fileCount := 0
		changeCount := 0

		for _, file := range files {
			changes, err := analyzeChanges(file)
			if err != nil {
				continue
			}

			if len(changes) > 0 {
				fmt.Printf("Would modify: %s\n", filepath.Base(file))
				for _, change := range changes {
					fmt.Printf("  - %s\n", change)
					changeCount++
				}
				fmt.Println()
				fileCount++
			}
		}

		if fileCount == 0 {
			fmt.Println("No files would be modified")
		} else {
			fmt.Printf("Would modify %d files (%d headings)\n", fileCount, changeCount)
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
		changes, _ := analyzeChanges(file)
		if len(changes) > 0 {
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

	fmt.Println("Removing IDs from headings...")
	fmt.Println()

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
