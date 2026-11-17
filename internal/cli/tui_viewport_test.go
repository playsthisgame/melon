package cli

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// 6 items × 2 rows = 12 rows needed; terminal is only 8 rows tall, leaving
	// 8 - listReservedRows (3) = 5 rows. The list must fit in those 5 rows as
	// rendered (bubbles needs height 4 here: one item plus the pagination row).
	m := newSearchModel(makeSearchItems(6))
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 8})
	sm := result.(searchModel)
	assert.LessOrEqual(t, sm.list.Height(), 5)
	assert.LessOrEqual(t, lipgloss.Height(sm.list.View()), 5)
}

func TestSearchModel_WindowSizeMsg_LargeTerminal(t *testing.T) {
	// 3 items × 2 rows = 6 rows initial height; terminal is 40 rows tall.
	// A tall terminal must not expand the list past its content-capped height.
	m := newSearchModel(makeSearchItems(3))
	initialHeight := m.list.Height()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	sm := result.(searchModel)
	assert.Equal(t, initialHeight, sm.list.Height())
}

// removeModel viewport tests

func TestRemoveModel_WindowSizeMsg_SmallTerminal(t *testing.T) {
	// 10 items × 1 row = 10 rows needed; terminal is 7 rows tall, leaving
	// 7 - listReservedRows (3) = 4 rows. The list must fit in those 4 rows as
	// rendered, pagination row included.
	m := newRemoveModel(makeRemoveItems(10))
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 7})
	rm := result.(removeModel)
	assert.LessOrEqual(t, rm.list.Height(), 4)
	assert.LessOrEqual(t, lipgloss.Height(rm.list.View()), 4)
}

func TestRemoveModel_WindowSizeMsg_LargeTerminal(t *testing.T) {
	// 3 items × 1 row = 3 rows initial height; terminal is 40 rows tall.
	// A tall terminal must not expand the list past its content-capped height.
	m := newRemoveModel(makeRemoveItems(3))
	initialHeight := m.list.Height()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	rm := result.(removeModel)
	assert.Equal(t, initialHeight, rm.list.Height())
}

// visibleCursor simulates bubbletea's standard renderer, which drops lines from
// the top of a view taller than the terminal, and reports whether the "> "
// cursor survives. It returns the number of rendered lines too.
func visibleCursor(view string, termHeight int) (lines int, visible bool) {
	all := strings.Split(view, "\n")
	shown := all
	if len(shown) > termHeight {
		shown = shown[len(shown)-termHeight:]
	}
	for _, l := range shown {
		if strings.Contains(l, "> ") {
			return len(all), true
		}
	}
	return len(all), false
}

type tuiModel interface {
	tea.Model
	View() string
}

// Regression for #49: in a short (split) terminal the view was taller than the
// terminal, so the cursor on the first rows was cut off.
func TestTUIModels_FitTerminalAndKeepCursorVisible(t *testing.T) {
	builders := map[string]func(n int) tuiModel{
		"search": func(n int) tuiModel { return newSearchModel(makeSearchItems(n)) },
		"remove": func(n int) tuiModel { return newRemoveModel(makeRemoveItems(n)) },
		"update": func(n int) tuiModel {
			items := make([]updateSkillItem, n)
			for i := range items {
				items[i] = updateSkillItem{name: "github.com/owner/skill", constraint: "^1.0.0"}
			}
			return newUpdateModel(items, nil)
		},
	}
	for name, build := range builders {
		for _, n := range []int{1, 2, 3, 6, 12, 30} {
			for h := 6; h <= 40; h++ {
				for _, down := range []int{0, n / 2, n - 1} {
					var m tea.Model = build(n)
					m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: h})
					for range down {
						m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
					}
					lines, ok := visibleCursor(m.View(), h)
					// Below 7 rows a 2-line search item plus the pagination row
					// cannot fit at all; the cursor must still stay visible.
					if h >= 7 {
						assert.LessOrEqual(t, lines, h, "%s n=%d h=%d down=%d: view taller than terminal", name, n, h, down)
					}
					assert.True(t, ok, "%s n=%d h=%d down=%d: cursor not visible", name, n, h, down)
				}
			}
		}
	}
}

func TestSearchModel_ShrinkWithCursorLowKeepsCursorVisible(t *testing.T) {
	var m tea.Model = newSearchModel(makeSearchItems(30))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	for range 9 {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	lines, ok := visibleCursor(m.View(), 10)
	assert.LessOrEqual(t, lines, 10)
	assert.True(t, ok, "cursor must stay visible after shrinking")
	assert.Equal(t, 9, m.(searchModel).list.Index())
}

func TestSearchModel_GrowsBackAfterResize(t *testing.T) {
	m := newSearchModel(makeSearchItems(30))
	initial := m.list.Height()
	r, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	r, _ = r.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	assert.Equal(t, initial, r.(searchModel).list.Height(), "list grows back to its content-capped height")
}
