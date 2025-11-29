package zettel

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupOptions configures backup behavior
type BackupOptions struct {
	BackupDir string // Directory where backups are stored
	Timestamp string // Timestamp for backup (defaults to current time)
}

// CreateBackup creates a backup of all markdown files in the zettel directory
func CreateBackup(zettelDir string, opts BackupOptions) (string, error) {
	// Generate timestamp if not provided
	if opts.Timestamp == "" {
		opts.Timestamp = time.Now().Format("20060102-150405")
	}

	// Determine backup directory
	backupDir := opts.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(zettelDir, ".backup-"+opts.Timestamp)
	}

	// Create backup directory
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get all markdown files
	files, err := GetMarkdownFiles(zettelDir)
	if err != nil {
		return "", fmt.Errorf("failed to get markdown files: %w", err)
	}

	// Copy each file to backup directory
	for _, file := range files {
		basename := filepath.Base(file)
		destPath := filepath.Join(backupDir, basename)

		err := copyFile(file, destPath)
		if err != nil {
			return "", fmt.Errorf("failed to backup %s: %w", basename, err)
		}
	}

	return backupDir, nil
}

// RestoreFromBackup restores files from a backup directory
func RestoreFromBackup(backupDir, zettelDir string) error {
	// Get all files in backup directory
	files, err := filepath.Glob(filepath.Join(backupDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no markdown files found in backup directory")
	}

	// Restore each file
	for _, file := range files {
		basename := filepath.Base(file)
		destPath := filepath.Join(zettelDir, basename)

		err := copyFile(file, destPath)
		if err != nil {
			return fmt.Errorf("failed to restore %s: %w", basename, err)
		}
	}

	return nil
}

// FindLatestBackup finds the most recent backup directory in zettelDir
func FindLatestBackup(zettelDir string) (string, error) {
	// Look for .backup-* directories
	pattern := filepath.Join(zettelDir, ".backup-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to search for backups: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no backups found")
	}

	// Find the most recent (lexicographically last due to timestamp format)
	var latest string
	for _, match := range matches {
		if latest == "" || match > latest {
			latest = match
		}
	}

	return latest, nil
}

// GetBackupTimestamp extracts the timestamp from a backup directory name
func GetBackupTimestamp(backupDir string) (time.Time, error) {
	basename := filepath.Base(backupDir)

	// Remove .backup- prefix
	if len(basename) < 8 || basename[:8] != ".backup-" {
		return time.Time{}, fmt.Errorf("invalid backup directory name")
	}

	timestamp := basename[8:]

	// Parse timestamp
	t, err := time.Parse("20060102-150405", timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	return t, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

// AskForConfirmation prompts the user for yes/no confirmation
func AskForConfirmation(prompt string) bool {
	var response string
	fmt.Printf("%s [y/N]: ", prompt)
	fmt.Scanln(&response)

	response = filepath.Clean(response) // Just to use filepath to avoid unused import

	if len(response) > 0 && (response[0] == 'y' || response[0] == 'Y') {
		return true
	}
	return false
}
