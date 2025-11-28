package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontMatter represents the YAML front-matter structure
type FrontMatter struct {
	ID string `yaml:"ID"`
}

// extractFrontMatter reads YAML front-matter from a markdown file
func extractFrontMatter(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var frontMatterLines []string
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
				// End of front-matter
				break
			}
		}

		if inFrontMatter && frontMatterCount == 1 {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	if frontMatterCount < 2 {
		return "", fmt.Errorf("no valid YAML front-matter found")
	}

	yamlContent := strings.Join(frontMatterLines, "\n")

	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	return fm.ID, nil
}

// getIDFromFilename extracts the ID from the filename (without .md extension)
func getIDFromFilename(filename string) string {
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, ".md")
}

// verifyFile checks if the ID in front-matter matches the filename
func verifyFile(filePath string) error {
	frontMatterID, err := extractFrontMatter(filePath)
	if err != nil {
		return err
	}

	filenameID := getIDFromFilename(filePath)

	if frontMatterID != filenameID {
		return fmt.Errorf("ID mismatch: front-matter has '%s' but filename is '%s'", frontMatterID, filenameID)
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

	hasErrors := false
	verifiedCount := 0

	for _, file := range files {
		if err := verifyFile(file); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", filepath.Base(file), err)
			hasErrors = true
		} else {
			fmt.Printf("✓ %s\n", filepath.Base(file))
			verifiedCount++
		}
	}

	fmt.Printf("\nVerified %d files in %s\n", verifiedCount, zettelDir)

	if hasErrors {
		os.Exit(1)
	}
}
