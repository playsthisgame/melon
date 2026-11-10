package resolver

import (
	"errors"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func byName(deps []ResolvedDep) map[string]ResolvedDep {
	out := make(map[string]ResolvedDep, len(deps))
	for _, d := range deps {
		out[d.Name] = d
	}
	return out
}

func TestResolve_PackMembersCarryProvenance(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{"github.com/community/spring-pack": "^1.0.0"}}
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/community/spring-pack|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{
			"github.com/acme/spring-boot": "^1.0.0",
			"github.com/acme/spring-data": "^1.0.0",
		}},
	})

	deps, err := Resolve(root, staticVersionResolver(nil), fetch)
	require.NoError(t, err)
	got := byName(deps)
	require.Len(t, got, 3)

	pack := got["github.com/community/spring-pack"]
	assert.Equal(t, manifest.TypePack, pack.Type)
	assert.True(t, pack.IsPack())
	assert.True(t, pack.Direct)
	assert.Nil(t, pack.Via)

	boot := got["github.com/acme/spring-boot"]
	assert.Equal(t, manifest.TypeSkill, boot.Type)
	assert.False(t, boot.Direct)
	assert.Equal(t, []string{"github.com/community/spring-pack"}, boot.Via)
}

func TestResolve_DiamondRecordsBothParents(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{
		"github.com/a/pack-b": "^1.0.0",
		"github.com/a/pack-a": "^1.0.0",
	}}
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/a/pack-a|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/x/junit": "^1.0.0"}},
		"https://github.com/a/pack-b|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/x/junit": "^1.0.0"}},
	})

	deps, err := Resolve(root, staticVersionResolver(nil), fetch)
	require.NoError(t, err)
	junit := byName(deps)["github.com/x/junit"]
	assert.False(t, junit.Direct)
	assert.Equal(t, []string{"github.com/a/pack-a", "github.com/a/pack-b"}, junit.Via)
}

func TestResolve_DirectAndTransitive(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{
		"github.com/community/spring-pack": "^1.0.0",
		"github.com/acme/spring-boot":      "^1.0.0",
	}}
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/community/spring-pack|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/acme/spring-boot": "^1.0.0"}},
	})

	deps, err := Resolve(root, staticVersionResolver(nil), fetch)
	require.NoError(t, err)
	boot := byName(deps)["github.com/acme/spring-boot"]
	assert.True(t, boot.Direct)
	assert.Equal(t, []string{"github.com/community/spring-pack"}, boot.Via)
}

func TestResolve_NestedPacks(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{"github.com/p/outer": "^1.0.0"}}
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/p/outer|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/p/inner": "^1.0.0"}},
		"https://github.com/p/inner|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/s/skill": "^1.0.0"}},
	})

	deps, err := Resolve(root, staticVersionResolver(nil), fetch)
	require.NoError(t, err)
	got := byName(deps)
	require.Len(t, got, 3)
	assert.Equal(t, []string{"github.com/p/outer"}, got["github.com/p/inner"].Via)
	assert.Equal(t, []string{"github.com/p/inner"}, got["github.com/s/skill"].Via)
}

func TestResolve_DirectOverrideCompatiblePicksDirectVersion(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{
		"github.com/community/spring-pack": "^1.0.0",
		"github.com/acme/spring-boot":      "^1.5.0",
	}}
	versions := staticVersionResolver(map[string]string{
		"https://github.com/acme/spring-boot|^1.5.0": "1.6.0",
		"https://github.com/acme/spring-boot|^1.0.0": "1.6.0",
	})
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/community/spring-pack|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/acme/spring-boot": "^1.0.0"}},
	})

	deps, err := Resolve(root, versions, fetch)
	require.NoError(t, err)
	assert.Equal(t, "1.6.0", byName(deps)["github.com/acme/spring-boot"].Version)
}

func TestResolve_DirectOverrideIncompatibleNamesPack(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{
		"github.com/community/spring-pack": "^1.0.0",
		"github.com/acme/spring-boot":      "^2.0.0",
	}}
	versions := staticVersionResolver(map[string]string{
		"https://github.com/acme/spring-boot|^2.0.0": "2.1.0",
	})
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/community/spring-pack|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/acme/spring-boot": "^1.0.0"}},
	})

	_, err := Resolve(root, versions, fetch)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrConflict))
	msg := err.Error()
	assert.Contains(t, msg, "melon.yaml")
	assert.Contains(t, msg, "pack github.com/community/spring-pack")
	assert.Contains(t, msg, "github.com/acme/spring-boot")
	assert.Contains(t, msg, `"^1.0.0"`)
	assert.Contains(t, msg, `"^2.0.0"`)
	assert.Contains(t, msg, "loosen")
}

func TestResolve_ProvenanceDeterministic(t *testing.T) {
	root := manifest.Manifest{Dependencies: map[string]string{
		"github.com/a/pack-a": "^1.0.0",
		"github.com/a/pack-b": "^1.0.0",
	}}
	fetch := staticManifestFetcher(map[string]manifest.Manifest{
		"https://github.com/a/pack-a|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/x/junit": "^1.0.0", "github.com/x/mockito": "^1.0.0"}},
		"https://github.com/a/pack-b|v1.0.0": {Type: manifest.TypePack, Dependencies: map[string]string{"github.com/x/junit": "^1.0.0"}},
	})
	first, err := Resolve(root, staticVersionResolver(nil), fetch)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		again, err := Resolve(root, staticVersionResolver(nil), fetch)
		require.NoError(t, err)
		assert.Equal(t, first, again)
	}
}
