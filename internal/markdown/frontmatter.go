package markdown

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FrontMatterKeywords extracts keywords from YAML front-matter
func FrontMatterKeywords(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
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
				inFrontMatter = false
				break
			}
		}

		if inFrontMatter && frontMatterCount == 1 {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if frontMatterCount < 2 {
		return nil, fmt.Errorf("no valid YAML front-matter found")
	}

	// Extract keyword values (the items in the list)
	var keywords []string
	inKeywords := false
	for _, line := range frontMatterLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "keywords:") {
			inKeywords = true
		} else if inKeywords {
			// Check if this is still part of the keywords list (starts with -)
			if strings.HasPrefix(trimmed, "-") {
				// Extract the keyword value after the dash
				keyword := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				if keyword != "" {
					keywords = append(keywords, keyword)
				}
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "-") {
				// We've reached a new key, stop collecting keywords
				break
			}
		}
	}

	return keywords, nil
}

// ReadContentAfterFrontMatter reads all content lines after front-matter
func ReadContentAfterFrontMatter(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var contentLines []string
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
				inFrontMatter = false
				continue
			}
		}

		if !inFrontMatter && frontMatterCount >= 2 {
			contentLines = append(contentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return contentLines, nil
}
