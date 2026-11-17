package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// listReservedRows is the number of terminal rows consumed by chrome outside the
// list itself: title line, hint bar, and trailing newline. Both searchModel and
// removeModel use this to cap their list height to the available terminal space.
const listReservedRows = 3

// fitListToTerminal sizes l for a terminal termHeight rows tall. The list gets
// the rows left after listReservedRows, but never more than maxHeight (its
// content-capped height), so it grows back after a resize as well as shrinking.
//
// bubbles' per-page math does not always leave room for the pagination row, so
// the height is reduced until the rendered list really fits. Output taller than
// the terminal makes bubbletea drop lines from the top, which is where the
// cursor is. The cursor's item is re-selected so its page stays in view.
func fitListToTerminal(l *list.Model, termHeight, maxHeight int) {
	available := termHeight - listReservedRows
	if available < 1 {
		available = 1
	}
	h := min(available, maxHeight)
	l.SetHeight(h)
	for h > 1 && lipgloss.Height(l.View()) > available {
		h--
		l.SetHeight(h)
	}
	l.Select(l.Index())
}

// searchResultItem is a single entry in the search results list.
type searchResultItem struct {
	path        string
	author      string
	description string
	featured    bool
	pack        bool
}

func (s searchResultItem) FilterValue() string { return s.path }

// searchMultiSelectDelegate renders each search result as two lines with a checkbox:
//
//	> [✓] ★ github.com/owner/repo  (author)
//	       description text
type searchMultiSelectDelegate struct {
	selected map[int]bool
}

func (d searchMultiSelectDelegate) Height() int                             { return 2 }
func (d searchMultiSelectDelegate) Spacing() int                            { return 0 }
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

	author := ""
	if r.author != "" {
		author = "  (" + r.author + ")"
	}

	desc := r.description
	if desc == "" {
		desc = "(no description)"
	}
	descLine := "           " + desc

	// The [pack] label keeps its own color on every row so packs stand out
	// from skills; the rest of the line follows the row's state style.
	nameStyle, descStyle, label := dimItem, dimItem, packLabelDim
	if d.selected[index] {
		nameStyle, descStyle, label = selectedItem, hintStyle, packLabel
	} else if index == m.Index() {
		nameStyle, descStyle, label = normalItem, hintStyle, packLabel
	}

	nameLine := nameStyle.Render(cursor + check + star + r.path)
	if r.pack {
		nameLine += " " + label.Render("[pack]")
	}
	if author != "" {
		nameLine += nameStyle.Render(author)
	}
	fmt.Fprint(w, nameLine+"\n")
	fmt.Fprint(w, descStyle.Render(descLine))
}

var (
	// packLabel marks packs in search results (magenta, so it differs from the
	// green selection and cyan title colors used elsewhere).
	packLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	packLabelDim = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

// searchModel is a bubbletea model for the interactive search results list.
type searchModel struct {
	list      list.Model
	maxHeight int // content-capped list height; the list never grows past it
	sel       map[int]bool
	selected  []string
	quitting  bool
}

func newSearchModel(results []searchResultItem) searchModel {
	items := make([]list.Item, len(results))
	for i, r := range results {
		items[i] = r
	}

	// Height: 2 lines per item + 2 rows of bubbles list internal chrome.
	// The chrome overhead must be added on top of the item rows.
	// Cap at 10 visible items: 10*2 + 2 = 22.
	height := len(results)*2 + 2
	if height > 22 {
		height = 22
	}

	sel := map[int]bool{}
	l := list.New(items, searchMultiSelectDelegate{selected: sel}, 80, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return searchModel{list: l, maxHeight: height, sel: sel}
}

func (m searchModel) Init() tea.Cmd { return nil }

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fitListToTerminal(&m.list, msg.Height, m.maxHeight)
		return m, nil
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
