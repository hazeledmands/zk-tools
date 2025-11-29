package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// UnfinishedNote represents a note that needs attention
type UnfinishedNote struct {
	FileID      string
	FilePath    string
	Title       string
	Reason      string
	WordCount   int
	TODOLines   []string
	SortKey     string
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
	var todoLines []string

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

		// Extract the first H1 heading as title
		trimmed := strings.TrimSpace(line)
		if title == "" && strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimSpace(trimmed[2:])
		}

		// Check for TODO markers (case-insensitive)
		upperLine := strings.ToUpper(line)
		if strings.Contains(upperLine, "TODO") || strings.Contains(upperLine, "FIXME") || strings.Contains(upperLine, "XXX") {
			todoLines = append(todoLines, strings.TrimSpace(line))
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
		FileID:    fileID,
		FilePath:  filePath,
		Title:     title,
		WordCount: wordCount,
		TODOLines: todoLines,
		SortKey:   strings.ToLower(title),
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
		if basename == "00-index.md" || basename == "TODO.md" {
			continue
		}

		noteInfo, err := extractNoteInfo(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process %s: %v\n", filepath.Base(file), err)
			continue
		}

		// Determine if this note is unfinished
		reasons := []string{}

		if len(noteInfo.TODOLines) > 0 {
			reasons = append(reasons, fmt.Sprintf("Contains %d TODO marker(s)", len(noteInfo.TODOLines)))
		}

		if noteInfo.WordCount < 100 {
			reasons = append(reasons, fmt.Sprintf("Only %d words", noteInfo.WordCount))
		}

		if len(reasons) > 0 {
			noteInfo.Reason = strings.Join(reasons, ", ")
			unfinishedNotes = append(unfinishedNotes, *noteInfo)
		}
	}

	// Sort by title
	sort.Slice(unfinishedNotes, func(i, j int) bool {
		return unfinishedNotes[i].SortKey < unfinishedNotes[j].SortKey
	})

	// Write to TODO.md
	todoPath := filepath.Join(zettelDir, "TODO.md")
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

	// Group by reason type
	withTODO := []UnfinishedNote{}
	shortNotes := []UnfinishedNote{}
	both := []UnfinishedNote{}

	for _, note := range unfinishedNotes {
		hasTODO := len(note.TODOLines) > 0
		isShort := note.WordCount < 100

		if hasTODO && isShort {
			both = append(both, note)
		} else if hasTODO {
			withTODO = append(withTODO, note)
		} else if isShort {
			shortNotes = append(shortNotes, note)
		}
	}

	// Write notes with both issues
	if len(both) > 0 {
		fmt.Fprintln(writer, "## Short Notes with TODOs")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes that are both short and contain TODO markers:\n", len(both))
		fmt.Fprintln(writer)

		for _, note := range both {
			fmt.Fprintf(writer, "### [[%s#%s]]\n", note.FileID, note.Title)
			fmt.Fprintf(writer, "- **Word count**: %d\n", note.WordCount)
			fmt.Fprintf(writer, "- **TODO markers**: %d\n", len(note.TODOLines))
			if len(note.TODOLines) > 0 {
				fmt.Fprintln(writer)
				for _, todoLine := range note.TODOLines {
					fmt.Fprintf(writer, "  - `%s`\n", todoLine)
				}
			}
			fmt.Fprintln(writer)
		}
	}

	// Write notes with TODOs
	if len(withTODO) > 0 {
		fmt.Fprintln(writer, "## Notes with TODO Markers")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes containing TODO markers:\n", len(withTODO))
		fmt.Fprintln(writer)

		for _, note := range withTODO {
			fmt.Fprintf(writer, "### [[%s#%s]]\n", note.FileID, note.Title)
			fmt.Fprintf(writer, "- **Word count**: %d\n", note.WordCount)
			fmt.Fprintf(writer, "- **TODO markers**: %d\n", len(note.TODOLines))
			fmt.Fprintln(writer)
			for _, todoLine := range note.TODOLines {
				fmt.Fprintf(writer, "  - `%s`\n", todoLine)
			}
			fmt.Fprintln(writer)
		}
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

	fmt.Printf("✓ Created TODO list with %d unfinished notes (%d with TODOs, %d short, %d both) at %s\n",
		len(unfinishedNotes), len(withTODO), len(shortNotes), len(both), todoPath)
}
