package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// helpers

func makeSearchItems(n int) []searchResultItem {
	items := make([]searchResultItem, n)
	for i := range items {
		items[i] = searchResultItem{path: "github.com/owner/skill", description: "desc"}
	}
	return items
}

func makeRemoveItems(n int) []removeSkillItem {
	items := make([]removeSkillItem, n)
	for i := range items {
		items[i] = removeSkillItem{name: "github.com/owner/skill", version: "^1.0.0"}
	}
	return items
}

// searchModel viewport tests

func TestSearchModel_WindowSizeMsg_SmallTerminal(t *testing.T) {
	// 6 items × 2 rows = 12 rows needed; terminal is only 8 rows tall.
	// Expected height: 8 - listReservedRows (3) = 5, clamped to [2, 12] → 5.
	m := newSearchModel(makeSearchItems(6))
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 8})
	sm := result.(searchModel)
	assert.Equal(t, 5, sm.list.Height())
}

func TestSearchModel_WindowSizeMsg_LargeTerminal(t *testing.T) {
	// 3 items × 2 rows = 6 rows initial height; terminal is 40 rows tall.
	// A tall terminal must NOT expand the list (expansion causes viewport scroll to bottom).
	// Height should remain at the initial content-capped value of 6.
	m := newSearchModel(makeSearchItems(3))
	initialHeight := m.list.Height()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	sm := result.(searchModel)
	assert.Equal(t, initialHeight, sm.list.Height())
}

// removeModel viewport tests

func TestRemoveModel_WindowSizeMsg_SmallTerminal(t *testing.T) {
	// 10 items × 1 row = 10 rows needed; terminal is 7 rows tall.
	// Expected height: 7 - listReservedRows (3) = 4, clamped to [2, 10] → 4.
	m := newRemoveModel(makeRemoveItems(10))
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 7})
	rm := result.(removeModel)
	assert.Equal(t, 4, rm.list.Height())
}

func TestRemoveModel_WindowSizeMsg_LargeTerminal(t *testing.T) {
	// 3 items × 1 row = 3 rows initial height; terminal is 40 rows tall.
	// A tall terminal must NOT expand the list (expansion causes viewport scroll to bottom).
	// Height should remain at the initial content-capped value of 3.
	m := newRemoveModel(makeRemoveItems(3))
	initialHeight := m.list.Height()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	rm := result.(removeModel)
	assert.Equal(t, initialHeight, rm.list.Height())
}
