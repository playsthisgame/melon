package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
)

// matchesAllowedSources reports whether depPath satisfies at least one glob
// pattern in patterns. An empty patterns slice permits all sources.
func matchesAllowedSources(depPath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		// path.Match handles glob semantics (* matches within a single segment).
		// To support "github.com/org/*" matching "github.com/org/repo/sub/path"
		// we also check prefix matching when the pattern ends with /*.
		if matched, _ := path.Match(pattern, depPath); matched {
			return true
		}
		// Allow trailing /* to match any subpath (e.g. "github.com/org/*" matches
		// "github.com/org/repo/sub/dir").
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(depPath, prefix) {
				return true
			}
		}
	}
	return false
}

// checkSourcePolicy validates depPaths against the allowed_sources policy in m.
// Returns a non-nil error listing all blocked dependencies when any are blocked.
// Returns nil when no policy is configured or all paths are permitted.
func checkSourcePolicy(m manifest.Manifest, depPaths []string) error {
	if m.Policy == nil || len(m.Policy.AllowedSources) == 0 {
		return nil
	}
	var blocked []string
	for _, dep := range depPaths {
		if !matchesAllowedSources(dep, m.Policy.AllowedSources) {
			blocked = append(blocked, dep)
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	return fmt.Errorf("source policy violation: the following dependencies are not permitted by allowed_sources %v:\n%s",
		m.Policy.AllowedSources, "  - "+strings.Join(blocked, "\n  - "))
}

// checkResolvedPolicy validates every dep in the resolved graph — including
// pack members and other transitive deps — against allowed_sources. Blocked
// transitive deps are reported with the deps that pulled them in, so users can
// see which pack is responsible.
func checkResolvedPolicy(m manifest.Manifest, resolved []resolver.ResolvedDep) error {
	if m.Policy == nil || len(m.Policy.AllowedSources) == 0 {
		return nil
	}
	var blocked []string
	for _, dep := range resolved {
		if matchesAllowedSources(dep.Name, m.Policy.AllowedSources) {
			continue
		}
		line := dep.Name
		if len(dep.Via) > 0 {
			line += " (via " + strings.Join(dep.Via, ", ") + ")"
		}
		blocked = append(blocked, line)
	}
	if len(blocked) == 0 {
		return nil
	}
	return fmt.Errorf("source policy violation: the following dependencies are not permitted by allowed_sources %v:\n%s",
		m.Policy.AllowedSources, "  - "+strings.Join(blocked, "\n  - "))
}
