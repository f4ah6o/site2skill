// Package search provides types for documentation search operations.
package search

// SearchResult represents a single search result from a documentation file.
// It includes metadata about where the match was found, how many times it occurred,
// surrounding context lines, and the source information.
type SearchResult struct {
	// File is the relative path to the documentation file containing the match.
	File string `json:"file"`
	// Matches is the total number of keyword matches found in this file.
	Matches int `json:"matches"`
	// Contexts is a slice of context snippets, each showing matches with surrounding lines.
	// Each line is prefixed with "> " for matched lines and "  " for context lines.
	Contexts []string `json:"contexts"`
	// SourceURL is the original URL where the documentation was fetched from.
	SourceURL string `json:"source_url"`
	// FetchedAt is the ISO 3339 timestamp when the documentation was fetched.
	FetchedAt string `json:"fetched_at"`
}

// Frontmatter represents YAML frontmatter metadata extracted from a Markdown file.
// It contains document metadata added during the conversion process.
type Frontmatter struct {
	// Title is the document title from the frontmatter.
	Title string
	// SourceURL is the original URL where the document was fetched from.
	SourceURL string
	// FetchedAt is the ISO 3339 timestamp when the document was fetched.
	FetchedAt string
}

// SearchOptions contains configuration parameters for search operations.
// It specifies the search scope, query, result limits, and output format.
type SearchOptions struct {
	// SkillDir is the path to the skill directory containing the docs/ subdirectory to search.
	SkillDir string
	// Query is the search query string (space-separated keywords with OR logic).
	Query string
	// MaxResults limits the number of results returned (0 means unlimited).
	MaxResults int
	// JSONOutput specifies whether to format results as JSON instead of human-readable text.
	JSONOutput bool
}
