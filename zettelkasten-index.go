package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Heading represents a heading found in a markdown file
type Heading struct {
	Text     string // The heading text (without # symbols)
	FileID   string // The file ID (filename without .md)
	Level    int    // Heading level (number of # symbols)
	SortKey  string // Lowercase version for case-insensitive sorting
}

// HeadingGroup represents an H1 heading with its nested subheadings
type HeadingGroup struct {
	Main     Heading   // The H1 heading
	Children []Heading // H2+ subheadings that follow this H1
}

// extractHeadings reads all headings from a markdown file
func extractHeadings(filePath string) ([]Heading, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileID := getIDFromFilename(filePath)
	var headings []Heading

	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	frontMatterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Skip YAML front-matter
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

		if inFrontMatter {
			continue
		}

		// Check if line is a heading
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Count the number of # symbols
			level := 0
			for _, char := range trimmed {
				if char == '#' {
					level++
				} else {
					break
				}
			}

			// Extract heading text (remove # symbols and trim)
			headingText := strings.TrimSpace(trimmed[level:])
			if headingText != "" {
				headings = append(headings, Heading{
					Text:    headingText,
					FileID:  fileID,
					Level:   level,
					SortKey: strings.ToLower(headingText),
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return headings, nil
}

// getIDFromFilename extracts the ID from the filename (without .md extension)
func getIDFromFilename(filename string) string {
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, ".md")
}

// formatObsidianLink creates an Obsidian-style link to a heading
func formatObsidianLink(heading Heading) string {
	return fmt.Sprintf("[[%s#%s]]", heading.FileID, heading.Text)
}

// shouldSkipHeading returns true if the heading should be excluded from the index
func shouldSkipHeading(headingText string) bool {
	// Trim trailing colon and convert to lowercase for comparison
	lower := strings.ToLower(strings.TrimSuffix(headingText, ":"))
	skipList := []string{
		"see also",
		"source",
		"sources",
		"further reading",
		"read more",
		"links",
		"references",
	}

	for _, skip := range skipList {
		if lower == skip {
			return true
		}
	}
	return false
}

// buildHeadingGroups organizes headings into hierarchical groups
// H1 headings become group mains, H2+ headings become children of the preceding H1
func buildHeadingGroups(headings []Heading) []HeadingGroup {
	var groups []HeadingGroup
	var currentGroup *HeadingGroup

	for _, heading := range headings {
		if heading.Level == 1 {
			// Save previous group if it exists
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}
			// Start new group with this H1
			currentGroup = &HeadingGroup{
				Main:     heading,
				Children: []Heading{},
			}
		} else if currentGroup != nil {
			// Add subheading to current group
			currentGroup.Children = append(currentGroup.Children, heading)
		}
		// If we encounter H2+ before any H1, we skip them (they have no parent)
	}

	// Don't forget to add the last group
	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}

	return groups
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

	// Collect all heading groups from all files
	var allGroups []HeadingGroup
	for _, file := range files {
		// Skip the index file itself
		if filepath.Base(file) == "00-index.md" {
			continue
		}

		headings, err := extractHeadings(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process %s: %v\n", filepath.Base(file), err)
			continue
		}

		// Filter out headings that should be skipped
		var filteredHeadings []Heading
		for _, heading := range headings {
			if !shouldSkipHeading(heading.Text) {
				filteredHeadings = append(filteredHeadings, heading)
			}
		}

		// Build hierarchical groups for this file
		groups := buildHeadingGroups(filteredHeadings)
		allGroups = append(allGroups, groups...)
	}

	// Sort groups alphabetically by the main heading's title (case-insensitive)
	sort.Slice(allGroups, func(i, j int) bool {
		return allGroups[i].Main.SortKey < allGroups[j].Main.SortKey
	})

	// Write to 00-index.md
	indexPath := filepath.Join(zettelDir, "00-index.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating index file: %v\n", err)
		os.Exit(1)
	}
	defer indexFile.Close()

	writer := bufio.NewWriter(indexFile)

	// Write header
	fmt.Fprintln(writer, "# Zettelkasten Index")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "All headings from all notes, sorted alphabetically:")
	fmt.Fprintln(writer)

	// Write all heading groups with level-based indentation
	totalCount := 0
	for _, group := range allGroups {
		// Write main heading (H1) - no indentation
		fmt.Fprintln(writer, formatObsidianLink(group.Main))
		totalCount++

		// Write children with indentation based on heading level
		// H2 = 2 spaces, H3 = 4 spaces, etc.
		for _, child := range group.Children {
			indent := strings.Repeat(" ", (child.Level-1)*2)
			fmt.Fprintf(writer, "%s%s\n", indent, formatObsidianLink(child))
			totalCount++
		}
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing index file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created index with %d headings (%d main, %d sub) at %s\n",
		totalCount, len(allGroups), totalCount-len(allGroups), indexPath)
}
