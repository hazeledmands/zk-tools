package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	zettelDir := os.Getenv("ZETTEL_DIR")
	if zettelDir == "" {
		zettelDir = filepath.Join(os.Getenv("HOME"), "Projects", "Zettelkasten")
	}

	if len(os.Args) > 1 {
		zettelDir = os.Args[1]
	}

	// Generate filename from current time: YYYYMMDDHHMM.md
	now := time.Now()
	filename := now.Format("200601021504") + ".md"
	filepath := filepath.Join(zettelDir, filename)

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Output the filename to stdout
	fmt.Println(filename)
}
