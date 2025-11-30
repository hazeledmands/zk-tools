package markdown

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CountLinksInLine counts the number of [[...]] links in a line
func CountLinksInLine(line string) int {
	count := 0
	start := 0
	for {
		idx := strings.Index(line[start:], "[[")
		if idx == -1 {
			break
		}
		idx += start

		// Find the closing ]]
		endIdx := strings.Index(line[idx:], "]]")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		count++
		start = endIdx + 2
	}
	return count
}

// ExtractLinksFromLine extracts all [[...]] link targets from a line
func ExtractLinksFromLine(line string) []string {
	var links []string
	start := 0
	for {
		idx := strings.Index(line[start:], "[[")
		if idx == -1 {
			break
		}
		idx += start

		// Find the closing ]]
		endIdx := strings.Index(line[idx:], "]]")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		// Extract the link content
		linkContent := line[idx+2 : endIdx]

		// Extract just the file ID (before # if present)
		fileID := linkContent
		if hashIdx := strings.Index(linkContent, "#"); hashIdx != -1 {
			fileID = linkContent[:hashIdx]
		}

		if fileID != "" {
			links = append(links, fileID)
		}

		start = endIdx + 2
	}
	return links
}

// ExtractHeadingLinksFromLine extracts all [[...#...]] heading links from a line
// Returns a map of "fileID#headingText" -> count
func ExtractHeadingLinksFromLine(line string) map[string]int {
	linkCounts := make(map[string]int)
	start := 0
	for {
		idx := strings.Index(line[start:], "[[")
		if idx == -1 {
			break
		}
		idx += start

		// Find the closing ]]
		endIdx := strings.Index(line[idx:], "]]")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		// Extract the link content
		linkContent := line[idx+2 : endIdx]

		// Check if this is a heading link (contains #)
		if hashIdx := strings.Index(linkContent, "#"); hashIdx != -1 {
			fileID := linkContent[:hashIdx]
			headingText := linkContent[hashIdx+1:]

			if fileID != "" && headingText != "" {
				key := fileID + "#" + headingText
				linkCounts[key]++
			}
		}

		start = endIdx + 2
	}
	return linkCounts
}

// CountBacklinks scans all markdown files and counts how many times each file is linked to
func CountBacklinks(files []string) (map[string]int, error) {
	backlinkCounts := make(map[string]int)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			links := ExtractLinksFromLine(line)
			for _, link := range links {
				backlinkCounts[link]++
			}
		}

		f.Close()
	}

	return backlinkCounts, nil
}

// CountHeadingBacklinks scans all markdown files and counts backlinks to specific headings
// Returns a map of "fileID#headingText" -> count
func CountHeadingBacklinks(files []string) (map[string]int, error) {
	backlinkCounts := make(map[string]int)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			linkCounts := ExtractHeadingLinksFromLine(line)
			for link, count := range linkCounts {
				backlinkCounts[link] += count
			}
		}

		f.Close()
	}

	return backlinkCounts, nil
}

// ExtractNoteInfo reads a markdown file and extracts relevant information
func ExtractNoteInfo(filePath string) (*NoteInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileID := GetIDFromFilename(filePath)
	scanner := bufio.NewScanner(file)

	inFrontMatter := false
	frontMatterCount := 0
	wordCount := 0
	outgoingLinkCount := 0
	var title string
	var todoHeadings []Heading
	var lastHeading *Heading

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

		trimmed := strings.TrimSpace(line)

		// Check if line is a heading
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
					// Extract the first H1 heading as title
					if title == "" && level == 1 {
						title = headingText
					}

					// Track this heading for potential TODO association
					lastHeading = &Heading{
						Text:   headingText,
						FileID: fileID,
						Level:  level,
					}
				}
			}
		} else if strings.HasPrefix(trimmed, "- #") || strings.HasPrefix(trimmed, "* #") {
			// This is a list item with a hashtag - extract keyword
			// Remove the list marker (- or *)
			content := strings.TrimSpace(trimmed[1:])

			// Extract keyword(s) from the line
			keywords := ExtractKeywordsFromLine(content)

			// Check if any keyword is "todo"
			for _, keyword := range keywords {
				if keyword == "todo" && lastHeading != nil {
					todoHeadings = append(todoHeadings, *lastHeading)
					break
				}
			}
		}

		// Count words (simple word count - split by whitespace)
		words := strings.Fields(line)
		wordCount += len(words)

		// Count outgoing links
		outgoingLinkCount += CountLinksInLine(line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// If no title found, use filename
	if title == "" {
		title = fileID
	}

	return &NoteInfo{
		FileID:            fileID,
		FilePath:          filePath,
		Title:             title,
		WordCount:         wordCount,
		TODOHeadings:      todoHeadings,
		OutgoingLinkCount: outgoingLinkCount,
		SortKey:           strings.ToLower(title),
	}, nil
}
