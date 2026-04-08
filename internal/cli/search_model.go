package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// searchResultItem is a single entry in the search results list.
type searchResultItem struct {
	path        string
	author      string
	description string
	featured    bool
}

func (s searchResultItem) FilterValue() string { return s.path }

// searchMultiSelectDelegate renders each search result as two lines with a checkbox:
//
//	> [✓] ★ github.com/owner/repo  (author)
//	       description text
type searchMultiSelectDelegate struct {
	selected map[int]bool
}

func (d searchMultiSelectDelegate) Height() int                              { return 2 }
func (d searchMultiSelectDelegate) Spacing() int                             { return 0 }
func (d searchMultiSelectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d searchMultiSelectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	r := item.(searchResultItem)

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	check := "[ ] "
	if d.selected[index] {
		check = "[✓] "
	}

	star := "  "
	if r.featured {
		star = "★ "
	}

	nameLine := cursor + check + star + r.path
	if r.author != "" {
		nameLine += "  (" + r.author + ")"
	}

	desc := r.description
	if desc == "" {
		desc = "(no description)"
	}
	descLine := "           " + desc

	if d.selected[index] {
		fmt.Fprintln(w, selectedItem.Render(nameLine))
		fmt.Fprintln(w, hintStyle.Render(descLine))
	} else if index == m.Index() {
		fmt.Fprintln(w, normalItem.Render(nameLine))
		fmt.Fprintln(w, hintStyle.Render(descLine))
	} else {
		fmt.Fprintln(w, dimItem.Render(nameLine))
		fmt.Fprintln(w, dimItem.Render(descLine))
	}
}

// searchModel is a bubbletea model for the interactive search results list.
type searchModel struct {
	list     list.Model
	sel      map[int]bool
	selected []string
	quitting bool
}

func newSearchModel(results []searchResultItem) searchModel {
	items := make([]list.Item, len(results))
	for i, r := range results {
		items[i] = r
	}

	// Height: 2 lines per item, cap at 10 visible items (20 lines).
	height := len(results) * 2
	if height > 20 {
		height = 20
	}

	sel := map[int]bool{}
	l := list.New(items, searchMultiSelectDelegate{selected: sel}, 80, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return searchModel{list: l, sel: sel}
}

func (m searchModel) Init() tea.Cmd { return nil }

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeySpace:
			idx := m.list.Index()
			m.sel[idx] = !m.sel[idx]
			m.list.SetDelegate(searchMultiSelectDelegate{selected: m.sel})
			return m, nil
		case tea.KeyEnter:
			items := m.list.Items()
			for i, item := range items {
				if m.sel[i] {
					m.selected = append(m.selected, item.(searchResultItem).path)
				}
			}
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m searchModel) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Search results") + "\n")
	b.WriteString(m.list.View() + "\n")
	b.WriteString(hintStyle.Render("↑↓ navigate  space to toggle  enter to install  esc to cancel") + "\n")
	return b.String()
}
