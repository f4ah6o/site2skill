package fetcher

import (
	"net/url"
	"path/filepath"
	"testing"
)

func TestNewFetcher(t *testing.T) {
	baseDir := "/tmp/test_fetcher"
	f := New(baseDir)
	if f == nil {
		t.Error("New() should return non-nil fetcher")
	}
}

func TestGetFilePath(t *testing.T) {
	f := New("/tmp/output")

	tests := []struct {
		name     string
		urlStr   string
		wantPath string
	}{
		// getFilePath appends to the provided directory, it doesn't add "crawl/" itself.
		// In this test we pass f.outputDir as the directory, so expectations should be relative to that.
		{
			name:     "Simple path",
			urlStr:   "https://example.com/docs/page",
			wantPath: "example.com/docs/page.html",
		},
		{
			name:     "Root path",
			urlStr:   "https://example.com/",
			wantPath: "example.com/index.html",
		},
		{
			name:     "Path with extension",
			urlStr:   "https://example.com/image.png",
			wantPath: "example.com/image.png",
		},
		{
			name:     "Query parameter (ja)",
			urlStr:   "https://ai.google.dev/gemini-api/docs?hl=ja",
			wantPath: "ai.google.dev/gemini-api/docs_q_hl_ja.html",
		},
		{
			name:     "Query parameter (en)",
			urlStr:   "https://ai.google.dev/gemini-api/docs?hl=en",
			wantPath: "ai.google.dev/gemini-api/docs_q_hl_en.html",
		},
		{
			name:   "Multiple query parameters",
			urlStr: "https://example.com/search?q=test&page=1",
			// Note: URL query encoding sorts keys
			wantPath: "example.com/search_q_page_1_q_test.html",
		},
		{
			name:   "Query parameter with special chars",
			urlStr: "https://example.com/path?key=val/ue",
			// % is stripped
			wantPath: "example.com/path_q_key_val2Fue.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			got := f.getFilePath(f.outputDir, u)
			// Remove base dir from result for easier comparison
			relPath, err := filepath.Rel(f.outputDir, got)
			if err != nil {
				t.Fatalf("Failed to get relative path: %v", err)
			}

			if relPath != tt.wantPath {
				t.Errorf("getFilePath() = %q, want %q", relPath, tt.wantPath)
			}
		})
	}
}
