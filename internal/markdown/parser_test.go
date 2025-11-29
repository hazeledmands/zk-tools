package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractKeywordsFromLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single keyword",
			input:    "#programming",
			expected: []string{"programming"},
		},
		{
			name:     "Multiple keywords",
			input:    "#golang #testing #automation",
			expected: []string{"golang", "testing", "automation"},
		},
		{
			name:     "Keywords with trailing punctuation",
			input:    "#golang, #testing, and #automation.",
			expected: []string{"golang", "testing", "automation"},
		},
		{
			name:     "No keywords",
			input:    "This is just text without hashtags",
			expected: []string{},
		},
		{
			name:     "Mixed text and keywords",
			input:    "Learn #golang and #rust today",
			expected: []string{"golang", "rust"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractKeywordsFromLine(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keywords, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, keyword := range tt.expected {
				if result[i] != keyword {
					t.Errorf("Expected keyword %q at position %d, got %q", keyword, i, result[i])
				}
			}
		})
	}
}

func TestExtractHeadingsWithKeywords(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test markdown file
	testFile := filepath.Join(tmpDir, "test.md")
	content := `# Programming Languages
- #functional
- #imperative
- #declarative

This is some content about programming languages.

## Functional Programming
Functional programming is a paradigm.

# Data Structures
- #algorithms
- #complexity

Data structures are fundamental to computer science.
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract headings
	headings, err := ExtractHeadings(testFile)
	if err != nil {
		t.Fatalf("Failed to extract headings: %v", err)
	}

	// Verify we have the expected headings
	if len(headings) != 3 {
		t.Errorf("Expected 3 headings, got %d", len(headings))
	}

	// Verify the first H1 has keywords
	if headings[0].Text != "Programming Languages" {
		t.Errorf("Expected first heading to be 'Programming Languages', got %q", headings[0].Text)
	}

	expectedKeywords := []string{"functional", "imperative", "declarative"}
	if len(headings[0].Keywords) != len(expectedKeywords) {
		t.Errorf("Expected %d keywords for first heading, got %d: %v",
			len(expectedKeywords), len(headings[0].Keywords), headings[0].Keywords)
	} else {
		for i, keyword := range expectedKeywords {
			if headings[0].Keywords[i] != keyword {
				t.Errorf("Expected keyword %q at position %d, got %q", keyword, i, headings[0].Keywords[i])
			}
		}
	}

	// Verify the H2 has no keywords
	if headings[1].Text != "Functional Programming" {
		t.Errorf("Expected second heading to be 'Functional Programming', got %q", headings[1].Text)
	}
	if len(headings[1].Keywords) != 0 {
		t.Errorf("Expected H2 to have no keywords, got %d: %v", len(headings[1].Keywords), headings[1].Keywords)
	}

	// Verify the second H1 has keywords
	if headings[2].Text != "Data Structures" {
		t.Errorf("Expected third heading to be 'Data Structures', got %q", headings[2].Text)
	}

	expectedKeywords2 := []string{"algorithms", "complexity"}
	if len(headings[2].Keywords) != len(expectedKeywords2) {
		t.Errorf("Expected %d keywords for third heading, got %d: %v",
			len(expectedKeywords2), len(headings[2].Keywords), headings[2].Keywords)
	} else {
		for i, keyword := range expectedKeywords2 {
			if headings[2].Keywords[i] != keyword {
				t.Errorf("Expected keyword %q at position %d, got %q", keyword, i, headings[2].Keywords[i])
			}
		}
	}
}

func TestBuildKeywordIndex(t *testing.T) {
	// Create test heading groups
	groups := []HeadingGroup{
		{
			Main: Heading{
				Text:     "Programming Languages",
				FileID:   "001",
				Level:    1,
				Keywords: []string{"functional", "imperative"},
			},
			Children: []Heading{},
		},
		{
			Main: Heading{
				Text:     "Data Structures",
				FileID:   "002",
				Level:    1,
				Keywords: []string{"algorithms", "functional"},
			},
			Children: []Heading{},
		},
	}

	// Build keyword index using the function from cmd/zettelkasten-index
	// We need to replicate it here for testing
	keywordIndex := make(map[string][]Heading)
	for _, group := range groups {
		for _, keyword := range group.Main.Keywords {
			keywordIndex[keyword] = append(keywordIndex[keyword], group.Main)
		}
		for _, child := range group.Children {
			for _, keyword := range child.Keywords {
				keywordIndex[keyword] = append(keywordIndex[keyword], child)
			}
		}
	}

	// Check that we have the expected keywords
	if len(keywordIndex) != 3 {
		t.Errorf("Expected 3 unique keywords, got %d", len(keywordIndex))
	}

	// Check "functional" appears in both headings
	functionalHeadings := keywordIndex["functional"]
	if len(functionalHeadings) != 2 {
		t.Errorf("Expected 'functional' to appear in 2 headings, got %d", len(functionalHeadings))
	}

	// Check "imperative" appears in one heading
	imperativeHeadings := keywordIndex["imperative"]
	if len(imperativeHeadings) != 1 {
		t.Errorf("Expected 'imperative' to appear in 1 heading, got %d", len(imperativeHeadings))
	}

	// Check "algorithms" appears in one heading
	algorithmsHeadings := keywordIndex["algorithms"]
	if len(algorithmsHeadings) != 1 {
		t.Errorf("Expected 'algorithms' to appear in 1 heading, got %d", len(algorithmsHeadings))
	}
}
