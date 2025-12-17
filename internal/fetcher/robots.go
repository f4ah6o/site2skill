package fetcher

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/temoto/robotstxt"
)

// RobotsManager handles fetching and checking robots.txt rules.
type RobotsManager struct {
	robotsData *robotstxt.RobotsData
	client     *http.Client
	userAgent  string
	mu         sync.Mutex
	fetched    bool
	basePath   string // Added basePath field
}

// NewRobotsManager creates a new RobotsManager.
func NewRobotsManager(client *http.Client, userAgent string, basePath string) *RobotsManager {
	// Ensure basePath ends with slash if it's a directory like
	if basePath != "" && basePath != "/" && !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	if basePath == "/" {
		basePath = ""
	}

	return &RobotsManager{
		client:    client,
		userAgent: userAgent,
		basePath:  basePath, // Initialize basePath
	}
}

// FetchRobotsTxt fetches and parses the robots.txt from the given base URL.
func (r *RobotsManager) FetchRobotsTxt(rootURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.fetched {
		return nil
	}

	parsedURL, err := url.Parse(rootURL)
	if err != nil {
		return err
	}

	robotsURL := parsedURL.Scheme + "://" + parsedURL.Host + "/robots.txt"
	log.Printf("Fetching robots.txt from %s", robotsURL)

	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch robots.txt: %v. Assuming allow all.", err)
		r.fetched = true
		return nil // Fail open
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("robots.txt returned status %d. Assuming allow all.", resp.StatusCode)
		r.fetched = true
		return nil
	}

	robotsData, err := robotstxt.FromResponse(resp)
	if err != nil {
		log.Printf("Failed to parse robots.txt: %v. Assuming allow all.", err)
		r.fetched = true
		return nil
	}

	r.robotsData = robotsData
	r.fetched = true
	log.Println("Successfully parsed robots.txt")
	return nil
}

// IsAllowed checks if the given URL is allowed by robots.txt.
func (r *RobotsManager) IsAllowed(targetURL string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.fetched || r.robotsData == nil {
		return true // Default allowed if no robots.txt or failed fetch
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return true
	}

	// IsAllowed expects path and query
	path := u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	// If basePath is configured and path starts with it, strip it
	// to match robots.txt rules which usually are relative to the site root
	// presented in robots.txt
	if r.basePath != "" && strings.HasPrefix(path, r.basePath) {
		// e.g. path: /site/docs/ng, base: /site/
		// rel: docs/ng -> make it /docs/ng
		path = "/" + strings.TrimPrefix(path, r.basePath)
	}

	group := r.robotsData.FindGroup(r.userAgent)
	return group.Test(path)
}
