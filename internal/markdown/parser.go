package markdown

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractKeywordsFromLine extracts hashtags from a line of text
func ExtractKeywordsFromLine(line string) []string {
	var keywords []string
	words := strings.Fields(line)

	for _, word := range words {
		// Remove trailing punctuation
		word = strings.TrimRight(word, ".,;:!?")
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			// Remove the # prefix and normalize to lowercase
			keyword := strings.ToLower(word[1:])
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

// GetIDFromFilename extracts the ID from the filename (without .md extension)
func GetIDFromFilename(filename string) string {
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, ".md")
}

// MakeSortKey creates a normalized sort key by removing leading special characters
func MakeSortKey(text string) string {
	// Find the first letter or digit
	for i, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			// Start from the first alphanumeric character
			return strings.ToLower(text[i:])
		}
	}
	// If no alphanumeric characters found, use the whole text lowercased
	return strings.ToLower(text)
}

// ShouldSkipHeading returns true if the heading should be excluded from the index
func ShouldSkipHeading(headingText string) bool {
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

// SkipFrontMatter skips YAML front-matter in a scanner
// Returns true if front-matter was found and skipped
func SkipFrontMatter(scanner *bufio.Scanner, callback func(line string) bool) error {
	inFrontMatter := false
	frontMatterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Check for front-matter delimiter
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

		// If we're past front-matter, pass line to callback
		// If callback returns false, stop processing
		if !inFrontMatter && frontMatterCount >= 2 {
			if !callback(line) {
				return nil
			}
		}
	}

	return scanner.Err()
}

// ExtractHeadings reads all headings from a markdown file
func ExtractHeadings(filePath string) ([]Heading, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileID := GetIDFromFilename(filePath)
	var headings []Heading

	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	frontMatterCount := 0
	var lastH1Index *int // Track the index of the last H1 heading

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

			// Check if there's a space after the # symbols (valid heading)
			if len(trimmed) > level && (trimmed[level] == ' ' || trimmed[level] == '\t') {
				// Extract heading text (remove # symbols and trim)
				headingText := strings.TrimSpace(trimmed[level:])
				if headingText != "" {
					idx := len(headings)
					headings = append(headings, Heading{
						Text:     headingText,
						FileID:   fileID,
						Level:    level,
						SortKey:  strings.ToLower(headingText),
						Keywords: []string{},
					})

					// Track H1 headings for keyword association
					if level == 1 {
						lastH1Index = &idx
					}
				}
			}
		} else if strings.HasPrefix(trimmed, "- #") || strings.HasPrefix(trimmed, "* #") {
			// This is a list item with a hashtag - extract keyword
			// Remove the list marker (- or *)
			content := strings.TrimSpace(trimmed[1:])

			// Extract keyword(s) from the line
			keywords := ExtractKeywordsFromLine(content)

			// Associate keywords with the last H1 heading
			if lastH1Index != nil && len(keywords) > 0 {
				headings[*lastH1Index].Keywords = append(headings[*lastH1Index].Keywords, keywords...)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return headings, nil
}

// BuildHeadingGroups organizes headings into hierarchical groups
// H1 headings become group mains, H2+ headings become children of the preceding H1
func BuildHeadingGroups(headings []Heading) []HeadingGroup {
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

// FormatObsidianLink creates an Obsidian-style link to a heading
func FormatObsidianLink(heading Heading) string {
	return fmt.Sprintf("[[%s#%s]]", heading.FileID, heading.Text)
}
