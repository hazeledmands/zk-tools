package markdown

// Heading represents a heading found in a markdown file
type Heading struct {
	Text     string   // The heading text (without # symbols)
	FileID   string   // The file ID (filename without .md)
	Level    int      // Heading level (number of # symbols)
	SortKey  string   // Lowercase version for case-insensitive sorting
	Keywords []string // Keywords associated with this heading (hashtags from lists)
}

// HeadingGroup represents an H1 heading with its nested subheadings
type HeadingGroup struct {
	Main     Heading   // The H1 heading
	Children []Heading // H2+ subheadings that follow this H1
}

// NoteInfo represents information extracted from a note file
type NoteInfo struct {
	FileID            string
	FilePath          string
	Title             string
	WordCount         int
	TODOHeadings      []Heading
	BacklinkCount     int
	OutgoingLinkCount int
	SortKey           string
}
