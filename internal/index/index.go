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
// If GitHub refuses the token, Fetch retries once without it: GitHub answers
// 404 even for public content when the token is invalid or expired, and a bad
// token must not break the public index.
func Fetch(rawURL string) ([]Entry, error) {
	onGitHub := false
	if u, err := url.Parse(rawURL); err == nil {
		onGitHub = githubHosts[strings.ToLower(u.Hostname())]
	}
	token := ""
	if onGitHub {
		token = gh.Token()
	}

	resp, err := get(rawURL, token)
	if err != nil {
		return nil, err
	}
	if token != "" && isAccessStatus(resp.StatusCode) {
		if retry, rerr := get(rawURL, ""); rerr == nil {
			if retry.StatusCode == http.StatusOK {
				resp.Body.Close()
				resp = retry
			} else {
				retry.Body.Close()
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if onGitHub && isAccessStatus(resp.StatusCode) {
			return nil, &AccessError{URL: rawURL, Status: resp.Status, HadToken: token != ""}
		}
		return nil, fmt.Errorf("index: fetch %s: unexpected status %s", rawURL, resp.Status)
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

// get issues a GET for rawURL, sending token as a Bearer token when non-empty.
func get(rawURL, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("index: fetch %s: %w", rawURL, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index: fetch %s: %w", rawURL, err)
	}
	return resp, nil
}

// AccessError is returned when GitHub refuses to serve an index: 401, 403, or
// the 404 GitHub sends for a private repository the caller cannot see. It is
// an access problem on GitHub's side, not a melon failure, and callers can
// show the fix (a token) instead of a generic error.
type AccessError struct {
	URL      string // the URL that was fetched
	Status   string // e.g. "404 Not Found"
	HadToken bool   // whether a GitHub token was sent
}

func (e *AccessError) Error() string {
	if !e.HadToken {
		return fmt.Sprintf("index: fetch %s: %s (if this index is in a private repository, set GITHUB_TOKEN to a token that can read it)", e.URL, e.Status)
	}
	return fmt.Sprintf("index: fetch %s: %s (check the URL, and that GITHUB_TOKEN can read this repository)", e.URL, e.Status)
}

func isAccessStatus(status int) bool {
	return status == http.StatusNotFound || status == http.StatusUnauthorized || status == http.StatusForbidden
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
