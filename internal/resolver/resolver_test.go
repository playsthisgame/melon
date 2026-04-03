package resolver

import (
	"errors"
	"fmt"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staticVersionResolver returns predetermined versions keyed by "repoURL|constraint".
// Falls back to "1.0.0" for unknown keys so tests only need to specify what matters.
func staticVersionResolver(table map[string]string) func(string, string) (string, string, error) {
	return func(repoURL, constraint string) (string, string, error) {
		key := repoURL + "|" + constraint
		v, ok := table[key]
		if !ok {
			v = "1.0.0"
		}
		return v, "v" + v, nil
	}
}

// staticManifestFetcher returns predetermined manifests keyed by "repoURL|gitTag".
// Returns an empty manifest for unknown keys (simulates no mln.yaml).
func staticManifestFetcher(table map[string]manifest.Manifest) func(string, string, string) (manifest.Manifest, error) {
	return func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		key := repoURL + "|" + gitTag
		m, ok := table[key]
		if !ok {
			return manifest.Manifest{}, nil
		}
		return m, nil
	}
}

func TestResolve_DirectOnly(t *testing.T) {
	// Fixture: one direct dep, no transitive deps.
	m, err := manifest.Load("testdata/fixture-direct-only.yaml")
	require.NoError(t, err)

	resolveVersion := staticVersionResolver(map[string]string{
		"https://github.com/alice/pdf-skill|^1.0.0": "1.2.0",
	})
	fetchManifest := staticManifestFetcher(nil) // no transitive manifests

	deps, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)
	require.Len(t, deps, 1)

	assert.Equal(t, "github.com/alice/pdf-skill", deps[0].Name)
	assert.Equal(t, "1.2.0", deps[0].Version)
	assert.Equal(t, "v1.2.0", deps[0].GitTag)
	assert.Equal(t, "https://github.com/alice/pdf-skill", deps[0].RepoURL)
	assert.Equal(t, "", deps[0].Subdir)
	assert.Equal(t, "SKILL.md", deps[0].Entrypoint)
}

func TestResolve_TransitiveInclusion(t *testing.T) {
	// Fixture: pdf-skill depends on bob/base-utils transitively.
	m, err := manifest.Load("testdata/fixture-transitive.yaml")
	require.NoError(t, err)

	resolveVersion := staticVersionResolver(map[string]string{
		"https://github.com/alice/pdf-skill|^1.0.0":   "1.2.0",
		"https://github.com/bob/base-utils|^2.0.0":    "2.1.0",
	})
	fetchManifest := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/alice/pdf-skill|v1.2.0": {
			Dependencies: map[string]string{
				"github.com/bob/base-utils": "^2.0.0",
			},
		},
		// bob/base-utils has no transitive deps (implicit empty manifest)
	})

	deps, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)
	require.Len(t, deps, 2)

	names := []string{deps[0].Name, deps[1].Name}
	assert.Contains(t, names, "github.com/alice/pdf-skill")
	assert.Contains(t, names, "github.com/bob/base-utils")

	for _, d := range deps {
		switch d.Name {
		case "github.com/alice/pdf-skill":
			assert.Equal(t, "1.2.0", d.Version)
		case "github.com/bob/base-utils":
			assert.Equal(t, "2.1.0", d.Version)
		}
	}
}

func TestResolve_DiamondResolution(t *testing.T) {
	// Two direct deps both depend on shared/lib with compatible constraints.
	// shared/lib must appear exactly once in the result.
	m := manifest.Manifest{
		Dependencies: map[string]string{
			"github.com/alice/a-skill": "^1.0.0",
			"github.com/alice/b-skill": "^1.0.0",
		},
	}

	resolveVersion := staticVersionResolver(map[string]string{
		"https://github.com/alice/a-skill|^1.0.0":   "1.0.0",
		"https://github.com/alice/b-skill|^1.0.0":   "1.0.0",
		"https://github.com/shared/lib|^1.0.0":       "1.5.0",
	})
	fetchManifest := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/alice/a-skill|v1.0.0": {
			Dependencies: map[string]string{"github.com/shared/lib": "^1.0.0"},
		},
		"https://github.com/alice/b-skill|v1.0.0": {
			Dependencies: map[string]string{"github.com/shared/lib": "^1.2.0"},
		},
	})

	deps, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)

	// shared/lib must appear exactly once.
	var sharedCount int
	for _, d := range deps {
		if d.Name == "github.com/shared/lib" {
			sharedCount++
			assert.Equal(t, "1.5.0", d.Version, "shared/lib should be pinned to resolved version")
		}
	}
	assert.Equal(t, 1, sharedCount, "shared/lib must appear exactly once in resolved set")
	assert.Len(t, deps, 3) // a-skill, b-skill, shared/lib
}

func TestResolve_VersionConflict(t *testing.T) {
	// Fixture: pdf-skill needs shared/lib@^1.x, word-skill needs shared/lib@^2.x — incompatible.
	m, err := manifest.Load("testdata/fixture-conflict.yaml")
	require.NoError(t, err)

	resolveVersion := staticVersionResolver(map[string]string{
		"https://github.com/alice/pdf-skill|^1.0.0":  "1.0.0",
		"https://github.com/alice/word-skill|^1.0.0":  "1.0.0",
		"https://github.com/shared/lib|^1.0.0":        "1.5.0",
	})
	fetchManifest := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/alice/pdf-skill|v1.0.0": {
			Dependencies: map[string]string{"github.com/shared/lib": "^1.0.0"},
		},
		"https://github.com/alice/word-skill|v1.0.0": {
			Dependencies: map[string]string{"github.com/shared/lib": "^2.0.0"},
		},
	})

	_, err = Resolve(m, resolveVersion, fetchManifest)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrConflict),
		"expected ErrConflict, got: %v", err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "github.com/shared/lib", "error must name the conflicting dep")
	t.Logf("conflict error: %v", err)
}

func TestResolve_MissingRemoteManifest(t *testing.T) {
	// A dep with no remote mln.yaml is treated as having no transitive deps.
	m := manifest.Manifest{
		Dependencies: map[string]string{
			"github.com/alice/no-manifest-skill": "^1.0.0",
		},
	}

	resolveVersion := staticVersionResolver(map[string]string{
		"https://github.com/alice/no-manifest-skill|^1.0.0": "1.0.0",
	})
	// fetchManifest returns empty manifest (simulates 404)
	fetchManifest := staticManifestFetcher(nil)

	deps, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err, "missing remote mln.yaml must not cause an error")
	require.Len(t, deps, 1)
	assert.Equal(t, "github.com/alice/no-manifest-skill", deps[0].Name)
}

func TestResolve_Deterministic(t *testing.T) {
	// Same manifest + same mocks → same result every time.
	m := manifest.Manifest{
		Dependencies: map[string]string{
			"github.com/z/z-skill": "^1.0.0",
			"github.com/a/a-skill": "^1.0.0",
			"github.com/m/m-skill": "^1.0.0",
		},
	}

	resolveVersion := func(repoURL, constraint string) (string, string, error) {
		return "1.0.0", "v1.0.0", nil
	}
	fetchManifest := staticManifestFetcher(nil)

	deps1, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)
	deps2, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)

	require.Len(t, deps1, 3)
	for i := range deps1 {
		assert.Equal(t, deps1[i].Name, deps2[i].Name, "results must be deterministic")
	}

	// Result must be sorted alphabetically.
	assert.Equal(t, "github.com/a/a-skill", deps1[0].Name)
	assert.Equal(t, "github.com/m/m-skill", deps1[1].Name)
	assert.Equal(t, "github.com/z/z-skill", deps1[2].Name)
}

func TestResolve_EntrypointFromManifest(t *testing.T) {
	// If a dep's mln.yaml specifies a custom entrypoint, it must be used.
	m := manifest.Manifest{
		Dependencies: map[string]string{
			"github.com/alice/custom-skill": "^1.0.0",
		},
	}

	resolveVersion := func(repoURL, constraint string) (string, string, error) {
		return "1.0.0", "v1.0.0", nil
	}
	fetchManifest := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/alice/custom-skill|v1.0.0": {
			Entrypoint: "docs/SKILL.md",
		},
	})

	deps, err := Resolve(m, resolveVersion, fetchManifest)
	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, "docs/SKILL.md", deps[0].Entrypoint)
}

func TestResolve_FetchManifestError(t *testing.T) {
	// A non-404 error from fetchManifest must propagate.
	m := manifest.Manifest{
		Dependencies: map[string]string{
			"github.com/alice/flaky-skill": "^1.0.0",
		},
	}

	resolveVersion := func(repoURL, constraint string) (string, string, error) {
		return "1.0.0", "v1.0.0", nil
	}
	fetchManifest := func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		return manifest.Manifest{}, fmt.Errorf("network timeout")
	}

	_, err := Resolve(m, resolveVersion, fetchManifest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
}
