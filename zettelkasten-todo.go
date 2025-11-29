package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HeadingWithTODO represents a heading that has a #todo tag
type HeadingWithTODO struct {
	Text  string
	Level int
}

// UnfinishedNote represents a note that needs attention
type UnfinishedNote struct {
	FileID         string
	FilePath       string
	Title          string
	WordCount      int
	TODOHeadings   []HeadingWithTODO
	SortKey        string
}

// extractKeywordsFromLine extracts hashtags from a line of text
func extractKeywordsFromLine(line string) []string {
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

// extractNoteInfo reads a markdown file and extracts relevant information
func extractNoteInfo(filePath string) (*UnfinishedNote, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileID := strings.TrimSuffix(filepath.Base(filePath), ".md")
	scanner := bufio.NewScanner(file)

	inFrontMatter := false
	frontMatterCount := 0
	wordCount := 0
	var title string
	var todoHeadings []HeadingWithTODO
	var lastHeading *HeadingWithTODO

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
					lastHeading = &HeadingWithTODO{
						Text:  headingText,
						Level: level,
					}
				}
			}
		} else if strings.HasPrefix(trimmed, "- #") || strings.HasPrefix(trimmed, "* #") {
			// This is a list item with a hashtag - extract keyword
			// Remove the list marker (- or *)
			content := strings.TrimSpace(trimmed[1:])

			// Extract keyword(s) from the line
			keywords := extractKeywordsFromLine(content)

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
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// If no title found, use filename
	if title == "" {
		title = fileID
	}

	return &UnfinishedNote{
		FileID:       fileID,
		FilePath:     filePath,
		Title:        title,
		WordCount:    wordCount,
		TODOHeadings: todoHeadings,
		SortKey:      strings.ToLower(title),
	}, nil
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

	var unfinishedNotes []UnfinishedNote

	for _, file := range files {
		// Skip the index and TODO files
		basename := filepath.Base(file)
		if basename == "00-index.md" || basename == "01-todo.md" {
			continue
		}

		noteInfo, err := extractNoteInfo(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process %s: %v\n", filepath.Base(file), err)
			continue
		}

		// Determine if this note is unfinished
		hasTODO := len(noteInfo.TODOHeadings) > 0
		isShort := noteInfo.WordCount < 100

		if hasTODO || isShort {
			unfinishedNotes = append(unfinishedNotes, *noteInfo)
		}
	}

	// Sort by title
	sort.Slice(unfinishedNotes, func(i, j int) bool {
		return unfinishedNotes[i].SortKey < unfinishedNotes[j].SortKey
	})

	// Write to 01-todo.md
	todoPath := filepath.Join(zettelDir, "01-todo.md")
	todoFile, err := os.Create(todoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating TODO file: %v\n", err)
		os.Exit(1)
	}
	defer todoFile.Close()

	writer := bufio.NewWriter(todoFile)

	// Write header
	fmt.Fprintln(writer, "# TODO: Unfinished Notes")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Notes that need attention (%d total):\n", len(unfinishedNotes))
	fmt.Fprintln(writer)

	// Group by reason type (can overlap)
	withTODO := []UnfinishedNote{}
	shortNotes := []UnfinishedNote{}

	for _, note := range unfinishedNotes {
		if len(note.TODOHeadings) > 0 {
			withTODO = append(withTODO, note)
		}
		if note.WordCount < 100 {
			shortNotes = append(shortNotes, note)
		}
	}

	// Write notes with TODOs
	if len(withTODO) > 0 {
		fmt.Fprintln(writer, "## Notes with #todo Tags")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes containing #todo tags:\n", len(withTODO))
		fmt.Fprintln(writer)

		for _, note := range withTODO {
			fmt.Fprintf(writer, "- [[%s#%s]] (%d sections with #todo)\n",
				note.FileID, note.Title, len(note.TODOHeadings))
		}
		fmt.Fprintln(writer)
	}

	// Write short notes
	if len(shortNotes) > 0 {
		fmt.Fprintln(writer, "## Short Notes")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes with fewer than 100 words:\n", len(shortNotes))
		fmt.Fprintln(writer)

		for _, note := range shortNotes {
			fmt.Fprintf(writer, "- [[%s#%s]] (%d words)\n", note.FileID, note.Title, note.WordCount)
		}
		fmt.Fprintln(writer)
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing TODO file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created TODO list with %d unfinished notes (%d with #todo, %d short) at %s\n",
		len(unfinishedNotes), len(withTODO), len(shortNotes), todoPath)
}
