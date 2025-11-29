package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hazel/zk-tools/internal/markdown"
	"github.com/hazel/zk-tools/internal/zettel"
)

func main() {
	zettelDir := zettel.GetZettelDirFromArgs(os.Args)

	files, err := zettel.GetMarkdownFiles(zettelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding markdown files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No markdown files found in %s\n", zettelDir)
		os.Exit(0)
	}

	// Count backlinks across all files
	backlinkCounts, err := markdown.CountBacklinks(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error counting backlinks: %v\n", err)
		os.Exit(1)
	}

	var unfinishedNotes []markdown.NoteInfo

	for _, file := range files {
		// Skip the index and TODO files
		if zettel.ShouldSkipFile(file) {
			continue
		}

		noteInfo, err := markdown.ExtractNoteInfo(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process %s: %v\n", filepath.Base(file), err)
			continue
		}

		// Determine if this note is unfinished
		hasTODO := len(noteInfo.TODOHeadings) > 0
		isShort := noteInfo.WordCount < 100

		if hasTODO || isShort {
			// Add backlink count
			noteInfo.BacklinkCount = backlinkCounts[noteInfo.FileID]
			unfinishedNotes = append(unfinishedNotes, *noteInfo)
		}
	}

	// Sort by backlinks (descending), then by word count (ascending)
	sort.Slice(unfinishedNotes, func(i, j int) bool {
		if unfinishedNotes[i].BacklinkCount != unfinishedNotes[j].BacklinkCount {
			return unfinishedNotes[i].BacklinkCount > unfinishedNotes[j].BacklinkCount
		}
		return unfinishedNotes[i].WordCount < unfinishedNotes[j].WordCount
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
	withTODO := []markdown.NoteInfo{}
	shortNotes := []markdown.NoteInfo{}
	orphanNotes := []markdown.NoteInfo{}

	for _, note := range unfinishedNotes {
		if len(note.TODOHeadings) > 0 {
			withTODO = append(withTODO, note)
		}
		if note.WordCount < 100 {
			shortNotes = append(shortNotes, note)
		}
		if note.BacklinkCount == 0 && note.OutgoingLinkCount == 0 {
			orphanNotes = append(orphanNotes, note)
		}
	}

	// Write orphan notes
	if len(orphanNotes) > 0 {
		fmt.Fprintln(writer, "## Orphan Notes")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes with no incoming or outgoing links:\n", len(orphanNotes))
		fmt.Fprintln(writer)

		for _, note := range orphanNotes {
			fmt.Fprintf(writer, "- [[%s#%s]] (%d words)\n",
				note.FileID, note.Title, note.WordCount)
		}
		fmt.Fprintln(writer)
	}

	// Write notes with TODOs
	if len(withTODO) > 0 {
		fmt.Fprintln(writer, "## Notes with #todo Tags")
		fmt.Fprintln(writer)
		fmt.Fprintf(writer, "%d notes containing #todo tags:\n", len(withTODO))
		fmt.Fprintln(writer)

		for _, note := range withTODO {
			fmt.Fprintf(writer, "- [[%s#%s]] (%d sections with #todo, %d backlinks)\n",
				note.FileID, note.Title, len(note.TODOHeadings), note.BacklinkCount)
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
			fmt.Fprintf(writer, "- [[%s#%s]] (%d words, %d backlinks)\n",
				note.FileID, note.Title, note.WordCount, note.BacklinkCount)
		}
		fmt.Fprintln(writer)
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing TODO file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created TODO list with %d unfinished notes (%d orphans, %d with #todo, %d short) at %s\n",
		len(unfinishedNotes), len(orphanNotes), len(withTODO), len(shortNotes), todoPath)
}
