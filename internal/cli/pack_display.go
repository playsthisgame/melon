package cli

import (
	"fmt"
	"strings"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/resolver"
)

// depGraph is a parent → children view of a resolved or locked dependency
// set, built from each dep's Via list. It lets prompts say how many skills a
// pack really brings in, rather than counting the pack as one skill.
type depGraph struct {
	children map[string][]string
	isPack   map[string]bool
}

func lockGraph(lf lockfile.LockFile) depGraph {
	g := depGraph{children: map[string][]string{}, isPack: map[string]bool{}}
	for _, d := range lf.Dependencies {
		if d.IsPack() {
			g.isPack[d.Name] = true
		}
		for _, p := range d.Via {
			g.children[p] = append(g.children[p], d.Name)
		}
	}
	return g
}

func resolvedGraph(deps []resolver.ResolvedDep) depGraph {
	g := depGraph{children: map[string][]string{}, isPack: map[string]bool{}}
	for _, d := range deps {
		if d.IsPack() {
			g.isPack[d.Name] = true
		}
		for _, p := range d.Via {
			g.children[p] = append(g.children[p], d.Name)
		}
	}
	return g
}

// skillsUnder returns every non-pack dep reachable from root, following nested
// packs and skills' own transitive deps.
func (g depGraph) skillsUnder(root string) map[string]bool {
	skills := map[string]bool{}
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, c := range g.children[n] {
			if seen[c] {
				continue
			}
			seen[c] = true
			if !g.isPack[c] {
				skills[c] = true
			}
			queue = append(queue, c)
		}
	}
	return skills
}

// unreachableAfter returns the locked deps that nothing would require once
// melon.yaml contains only remaining — i.e. what a full install would prune.
// Entries from lock files written before provenance existed (no type) are
// treated as roots unless they are being removed, since their parents are
// unknown.
func unreachableAfter(lf lockfile.LockFile, remaining map[string]string, removing map[string]bool) map[string]bool {
	g := lockGraph(lf)
	reach := map[string]bool{}
	var queue []string
	for _, d := range lf.Dependencies {
		_, direct := remaining[d.Name]
		legacyRoot := d.Type == "" && !removing[d.Name]
		if direct || legacyRoot {
			reach[d.Name] = true
			queue = append(queue, d.Name)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, c := range g.children[n] {
			if !reach[c] {
				reach[c] = true
				queue = append(queue, c)
			}
		}
	}
	out := map[string]bool{}
	for _, d := range lf.Dependencies {
		if !reach[d.Name] {
			out[d.Name] = true
		}
	}
	return out
}

// packTag renders the colored "[pack · N skills]" label.
func packTag(skills int) string {
	return packLabel.Render(fmt.Sprintf("[pack · %s]", count(skills, "skill")))
}

// selectionSummary renders "1 pack (7 skills) and 2 skills" for prompts.
// packSkills is the number of distinct skills the selected packs account for.
func selectionSummary(packs, packSkills, plain int) string {
	var parts []string
	if packs > 0 {
		parts = append(parts, fmt.Sprintf("%s (%s)", count(packs, "pack"), count(packSkills, "skill")))
	}
	if plain > 0 {
		parts = append(parts, count(plain, "skill"))
	}
	return strings.Join(parts, " and ")
}
