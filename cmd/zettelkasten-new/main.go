package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hazel/zk-tools/internal/zettel"
)

func main() {
	zettelDir := zettel.GetZettelDirFromArgs(os.Args)

	// Generate filename from current time: YYYYMMDDHHMM.md
	now := time.Now()
	filename := now.Format("200601021504") + ".md"
	filePath := filepath.Join(zettelDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Output the filename to stdout
	fmt.Println(filename)
}
