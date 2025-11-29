package zettel

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test markdown files
	files := map[string]string{
		"note1.md": "# Note 1\nContent 1",
		"note2.md": "# Note 2\nContent 2",
		"note3.md": "# Note 3\nContent 3",
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create backup with custom timestamp
	opts := BackupOptions{
		Timestamp: "20240101-120000",
	}

	backupDir, err := CreateBackup(tmpDir, opts)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup directory was created
	expectedBackupDir := filepath.Join(tmpDir, ".backup-20240101-120000")
	if backupDir != expectedBackupDir {
		t.Errorf("Expected backup dir %q, got %q", expectedBackupDir, backupDir)
	}

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup directory was not created")
	}

	// Verify all files were backed up
	for filename, expectedContent := range files {
		backupPath := filepath.Join(backupDir, filename)
		content, err := os.ReadFile(backupPath)
		if err != nil {
			t.Errorf("Failed to read backup file %s: %v", filename, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Backup content mismatch for %s. Expected %q, got %q",
				filename, expectedContent, string(content))
		}
	}
}

func TestCreateBackup_DefaultTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Create backup with default timestamp
	backupDir, err := CreateBackup(tmpDir, BackupOptions{})
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup directory name has .backup- prefix
	basename := filepath.Base(backupDir)
	if len(basename) < 8 || basename[:8] != ".backup-" {
		t.Errorf("Backup directory should start with .backup-, got %q", basename)
	}

	// Verify timestamp format (should be parseable)
	_, err = GetBackupTimestamp(backupDir)
	if err != nil {
		t.Errorf("Backup timestamp should be valid: %v", err)
	}
}

func TestCreateBackup_CustomBackupDir(t *testing.T) {
	tmpDir := t.TempDir()
	customBackupDir := filepath.Join(tmpDir, "my-custom-backup")

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Create backup with custom directory
	opts := BackupOptions{
		BackupDir: customBackupDir,
	}

	backupDir, err := CreateBackup(tmpDir, opts)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if backupDir != customBackupDir {
		t.Errorf("Expected custom backup dir %q, got %q", customBackupDir, backupDir)
	}

	// Verify file was backed up
	backupFile := filepath.Join(customBackupDir, "test.md")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file not created in custom directory")
	}
}

func TestRestoreFromBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create original files
	originalFiles := map[string]string{
		"note1.md": "# Original Note 1",
		"note2.md": "# Original Note 2",
	}

	for filename, content := range originalFiles {
		path := filepath.Join(tmpDir, filename)
		os.WriteFile(path, []byte(content), 0644)
	}

	// Create backup
	backupDir, err := CreateBackup(tmpDir, BackupOptions{Timestamp: "20240101-120000"})
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify original files
	modifiedFiles := map[string]string{
		"note1.md": "# Modified Note 1",
		"note2.md": "# Modified Note 2",
	}

	for filename, content := range modifiedFiles {
		path := filepath.Join(tmpDir, filename)
		os.WriteFile(path, []byte(content), 0644)
	}

	// Verify files are modified
	content, _ := os.ReadFile(filepath.Join(tmpDir, "note1.md"))
	if string(content) != modifiedFiles["note1.md"] {
		t.Fatal("Setup failed: file should be modified")
	}

	// Restore from backup
	err = RestoreFromBackup(backupDir, tmpDir)
	if err != nil {
		t.Fatalf("Failed to restore from backup: %v", err)
	}

	// Verify files are restored to original content
	for filename, expectedContent := range originalFiles {
		path := filepath.Join(tmpDir, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", filename, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Restore failed for %s. Expected %q, got %q",
				filename, expectedContent, string(content))
		}
	}
}

func TestFindLatestBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple backup directories
	backupDirs := []string{
		".backup-20240101-120000",
		".backup-20240102-120000",
		".backup-20240103-120000",
	}

	for _, dir := range backupDirs {
		path := filepath.Join(tmpDir, dir)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			t.Fatalf("Failed to create backup dir: %v", err)
		}
	}

	// Find latest backup
	latest, err := FindLatestBackup(tmpDir)
	if err != nil {
		t.Fatalf("Failed to find latest backup: %v", err)
	}

	expectedLatest := filepath.Join(tmpDir, ".backup-20240103-120000")
	if latest != expectedLatest {
		t.Errorf("Expected latest backup %q, got %q", expectedLatest, latest)
	}
}

func TestFindLatestBackup_NoBackups(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindLatestBackup(tmpDir)
	if err == nil {
		t.Error("Expected error when no backups exist")
	}
}

func TestGetBackupTimestamp(t *testing.T) {
	tests := []struct {
		name          string
		backupDir     string
		expectedTime  string
		shouldError   bool
	}{
		{
			name:         "Valid timestamp",
			backupDir:    "/path/.backup-20240101-120000",
			expectedTime: "2024-01-01 12:00:00",
			shouldError:  false,
		},
		{
			name:         "Valid timestamp different date",
			backupDir:    ".backup-20231225-235959",
			expectedTime: "2023-12-25 23:59:59",
			shouldError:  false,
		},
		{
			name:        "Invalid format - missing prefix",
			backupDir:   "backup-20240101-120000",
			shouldError: true,
		},
		{
			name:        "Invalid format - bad timestamp",
			backupDir:   ".backup-invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := GetBackupTimestamp(tt.backupDir)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			expectedTime, _ := time.Parse("2006-01-02 15:04:05", tt.expectedTime)
			if !timestamp.Equal(expectedTime) {
				t.Errorf("Expected time %v, got %v", expectedTime, timestamp)
			}
		})
	}
}

func TestRestoreFromBackup_EmptyBackup(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, ".backup-20240101-120000")
	os.MkdirAll(backupDir, 0755)

	err := RestoreFromBackup(backupDir, tmpDir)
	if err == nil {
		t.Error("Expected error when restoring from empty backup")
	}
}
