package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hazel/zk-tools/internal/markdown"
	"github.com/hazel/zk-tools/internal/zettel"
)

// buildKeywordIndex creates a map from keywords to the headings that reference them
func buildKeywordIndex(groups []markdown.HeadingGroup) map[string][]markdown.Heading {
	keywordIndex := make(map[string][]markdown.Heading)

	for _, group := range groups {
		// Add keywords from the main H1 heading
		for _, keyword := range group.Main.Keywords {
			keywordIndex[keyword] = append(keywordIndex[keyword], group.Main)
		}

		// Add keywords from child headings
		for _, child := range group.Children {
			for _, keyword := range child.Keywords {
				keywordIndex[keyword] = append(keywordIndex[keyword], child)
			}
		}
	}

	return keywordIndex
}

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

	// Count backlinks to headings across all files
	backlinkCounts, err := markdown.CountHeadingBacklinks(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to count backlinks: %v\n", err)
	}

	// Collect all heading groups from all files
	var allGroups []markdown.HeadingGroup
	for _, file := range files {
		// Skip the index file itself
		if zettel.ShouldSkipFile(file, "00-index.md") {
			continue
		}

		headings, err := markdown.ExtractHeadings(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process %s: %v\n", filepath.Base(file), err)
			continue
		}

		// Filter out headings that should be skipped
		var filteredHeadings []markdown.Heading
		for _, heading := range headings {
			if !markdown.ShouldSkipHeading(heading.Text) {
				// Update backlink count for this heading
				backlinkKey := heading.FileID + "#" + heading.Text
				heading.BacklinkCount = backlinkCounts[backlinkKey]
				filteredHeadings = append(filteredHeadings, heading)
			}
		}

		// Build hierarchical groups for this file
		groups := markdown.BuildHeadingGroups(filteredHeadings)
		allGroups = append(allGroups, groups...)
	}

	// Sort groups alphabetically by the main heading's title (case-insensitive)
	sort.Slice(allGroups, func(i, j int) bool {
		return allGroups[i].Main.SortKey < allGroups[j].Main.SortKey
	})

	// Build keyword index
	keywordIndex := buildKeywordIndex(allGroups)

	// Get sorted list of keywords
	var keywords []string
	for keyword := range keywordIndex {
		keywords = append(keywords, keyword)
	}
	sort.Strings(keywords)

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

	// Write keyword index if there are any keywords
	if len(keywords) > 0 {
		fmt.Fprintln(writer, "## Keywords")
		fmt.Fprintln(writer)

		for _, keyword := range keywords {
			fmt.Fprintf(writer, "### #%s\n", keyword)
			fmt.Fprintln(writer)

			// Write all headings that reference this keyword
			headings := keywordIndex[keyword]
			for _, heading := range headings {
				fmt.Fprintf(writer, "- %s (words: %d, backlinks: %d)\n",
					markdown.FormatObsidianLink(heading),
					heading.WordCount,
					heading.BacklinkCount)
			}
			fmt.Fprintln(writer)
		}
	}

	fmt.Fprintln(writer, "## All Headings")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "All headings from all notes, sorted alphabetically:")
	fmt.Fprintln(writer)

	// Write all heading groups with level-based indentation
	totalCount := 0
	for _, group := range allGroups {
		// Write main heading (H1) - no indentation
		// Format: [[link]] (words: N, backlinks: N)
		fmt.Fprintf(writer, "%s (words: %d, backlinks: %d)\n",
			markdown.FormatObsidianLink(group.Main),
			group.Main.WordCount,
			group.Main.BacklinkCount)
		totalCount++

		// Write children with indentation based on heading level
		// H2 = 2 spaces, H3 = 4 spaces, etc.
		for _, child := range group.Children {
			indent := strings.Repeat(" ", (child.Level-1)*2)
			fmt.Fprintf(writer, "%s%s (words: %d, backlinks: %d)\n",
				indent,
				markdown.FormatObsidianLink(child),
				child.WordCount,
				child.BacklinkCount)
			totalCount++
		}
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing index file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created index with %d headings (%d main, %d sub) and %d keywords at %s\n",
		totalCount, len(allGroups), totalCount-len(allGroups), len(keywords), indexPath)
}
