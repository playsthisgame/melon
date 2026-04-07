package index

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const IndexURL = "https://raw.githubusercontent.com/playsthisgame/melon-index/main/index.yml"

// Entry is a single skill record from the melon-index.
type Entry struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Featured    bool     `yaml:"featured"`
}

type indexFile struct {
	Skills []Entry `yaml:"skills"`
}

// Fetch downloads and parses the melon-index index.yml.
func Fetch() ([]Entry, error) {
	resp, err := http.Get(IndexURL)
	if err != nil {
		return nil, fmt.Errorf("index: fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index: fetch: unexpected status %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("index: read: %w", err)
	}
	var idx indexFile
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("index: parse: %w", err)
	}
	return idx.Skills, nil
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
