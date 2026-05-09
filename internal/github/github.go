package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	modver "golang.org/x/mod/semver"
)

const defaultAPIBase = "https://api.github.com"

// Client is a thin GitHub REST API client. It reads GITHUB_TOKEN from the
// environment and sends it as a Bearer token when present.
type Client struct {
	http    *http.Client
	token   string
	apiBase string // defaults to https://api.github.com
}

// New returns a Client configured from the environment.
func New() *Client {
	return &Client{
		http:    &http.Client{},
		token:   os.Getenv("GITHUB_TOKEN"),
		apiBase: defaultAPIBase,
	}
}

// NewWithBase returns a Client that uses baseURL instead of https://api.github.com.
// Intended for testing.
func NewWithBase(baseURL string) *Client {
	return &Client{
		http:    &http.Client{},
		apiBase: baseURL,
	}
}

func (c *Client) do(apiURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("github: rate limit exceeded — set GITHUB_TOKEN to increase the limit (60 → 5000 req/hr)")
	}
	return resp, nil
}

// SearchResult is a repository returned by the GitHub search API.
type SearchResult struct {
	// Name is the installable GitHub path, e.g. "github.com/owner/repo".
	Name        string
	Description string
	Owner       string
}

type searchAPIResponse struct {
	Items []struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"items"`
}

// SearchByTopic queries the GitHub repository search API for repos tagged with
// the melon-skill topic that match term.
func (c *Client) SearchByTopic(term string) ([]SearchResult, error) {
	q := url.QueryEscape("topic:melon-skill " + term)
	resp, err := c.do(c.apiBase+"/search/repositories?q=" + q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: search: read: %w", err)
	}
	var sr searchAPIResponse
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, fmt.Errorf("github: search: parse: %w", err)
	}
	results := make([]SearchResult, len(sr.Items))
	for i, item := range sr.Items {
		results[i] = SearchResult{
			Name:        "github.com/" + item.FullName,
			Description: item.Description,
			Owner:       item.Owner.Login,
		}
	}
	return results, nil
}

// RepoMeta fetches the description (about field) for a GitHub repo.
func (c *Client) RepoMeta(owner, repo string) (string, error) {
	resp, err := c.do(fmt.Sprintf(c.apiBase+"/repos/%s/%s", owner, repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return "", fmt.Errorf("github: repo %s/%s not found", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github: repo meta: unexpected status %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("github: repo meta: read: %w", err)
	}
	var r struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("github: repo meta: parse: %w", err)
	}
	return r.Description, nil
}

// ListTags returns semver tags for owner/repo sorted descending (newest first).
func (c *Client) ListTags(owner, repo string) ([]string, error) {
	resp, err := c.do(fmt.Sprintf(c.apiBase+"/repos/%s/%s/tags?per_page=100", owner, repo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: list tags: unexpected status %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: list tags: read: %w", err)
	}
	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("github: list tags: parse: %w", err)
	}
	var semverTags []string
	for _, t := range tags {
		v := t.Name
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		if modver.IsValid(v) {
			semverTags = append(semverTags, t.Name)
		}
	}
	sort.Slice(semverTags, func(i, j int) bool {
		vi, vj := semverTags[i], semverTags[j]
		if !strings.HasPrefix(vi, "v") {
			vi = "v" + vi
		}
		if !strings.HasPrefix(vj, "v") {
			vj = "v" + vj
		}
		return modver.Compare(vi, vj) > 0 // descending: newest first
	})
	return semverTags, nil
}

// ListBranches returns the branch names for owner/repo.
func (c *Client) ListBranches(owner, repo string) ([]string, error) {
	resp, err := c.do(fmt.Sprintf(c.apiBase+"/repos/%s/%s/branches?per_page=100", owner, repo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: list branches: unexpected status %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: list branches: read: %w", err)
	}
	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &branches); err != nil {
		return nil, fmt.Errorf("github: list branches: parse: %w", err)
	}
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	return names, nil
}
