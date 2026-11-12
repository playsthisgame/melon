package index

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	gh "github.com/playsthisgame/melon/internal/github"
	"gopkg.in/yaml.v3"
)

const DefaultIndexURL = "https://raw.githubusercontent.com/playsthisgame/melon/main/index.yaml"

// Entry is a single skill record from the melon-index.
type Entry struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Featured    bool     `yaml:"featured"`
	// Type is "skill" (default) or "pack". Unknown values are treated as
	// "skill"; use IsPack rather than reading the field directly.
	Type string `yaml:"type,omitempty"`
}

// IsPack reports whether the entry is a pack (a curated group of skills).
func (e Entry) IsPack() bool {
	return strings.EqualFold(e.Type, "pack")
}

type indexFile struct {
	Skills []Entry `yaml:"skills"`
}

// githubHosts are the hosts that receive the GitHub token. The token is never
// sent anywhere else, so a custom index on another host cannot capture it.
var githubHosts = map[string]bool{
	"raw.githubusercontent.com": true,
	"api.github.com":            true,
	"github.com":                true,
}

// Fetch downloads and parses the index.yaml at the given URL. Requests to
// GitHub hosts carry the GitHub token when one is set, so indices in private
// repositories (including GitHub Enterprise Cloud organizations) can be read.
func Fetch(rawURL string) ([]Entry, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("index: fetch %s: %w", rawURL, err)
	}
	onGitHub := false
	if u, err := url.Parse(rawURL); err == nil {
		onGitHub = githubHosts[strings.ToLower(u.Hostname())]
	}
	token := ""
	if onGitHub {
		token = gh.Token()
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index: fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index: fetch %s: unexpected status %s%s",
			rawURL, resp.Status, accessHint(onGitHub, token != "", resp.StatusCode))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("index: read %s: %w", rawURL, err)
	}
	var idx indexFile
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("index: parse %s: %w", rawURL, err)
	}
	return idx.Skills, nil
}

// accessHint explains a failed GitHub fetch. GitHub answers 404 rather than
// 401/403 for a private repository the caller cannot see.
func accessHint(onGitHub, hadToken bool, status int) string {
	if !onGitHub {
		return ""
	}
	switch status {
	case http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden:
	default:
		return ""
	}
	if !hadToken {
		return " (if this index is in a private repository, set GITHUB_TOKEN to a token that can read it)"
	}
	return " (check the URL, and that GITHUB_TOKEN can read this repository)"
}

// Search filters entries by term (case-insensitive) against name, description,
// author, and tags. Featured entries are returned before non-featured.
func Search(entries []Entry, term string) []Entry {
	term = strings.ToLower(term)
	var featured, regular []Entry
	for _, e := range entries {
		if matches(e, term) {
			if e.Featured {
				featured = append(featured, e)
			} else {
				regular = append(regular, e)
			}
		}
	}
	return append(featured, regular...)
}

// Find returns the index entry for the given skill name, or nil if not found.
func Find(entries []Entry, name string) *Entry {
	name = strings.TrimRight(name, "/")
	for i, e := range entries {
		if strings.EqualFold(e.Name, name) {
			return &entries[i]
		}
	}
	return nil
}

func matches(e Entry, term string) bool {
	if strings.Contains(strings.ToLower(e.Name), term) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Description), term) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Author), term) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return true
		}
	}
	return false
}
