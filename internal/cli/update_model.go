package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const updateAllSentinel = "__UPDATE_ALL__"

// updateSkillItem is a single entry in the interactive update list.
type updateSkillItem struct {
	name       string
	constraint string
	locked     string // currently locked version, or "" if not yet locked
}

func (u updateSkillItem) FilterValue() string { return u.name }

// updateMultiSelectDelegate renders each dep as a single line with a checkbox.
type updateMultiSelectDelegate struct {
	selected map[int]bool
}

func (d updateMultiSelectDelegate) Height() int                              { return 1 }
func (d updateMultiSelectDelegate) Spacing() int                             { return 0 }
func (d updateMultiSelectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d updateMultiSelectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	u := item.(updateSkillItem)

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	check := "[ ] "
	if d.selected[index] {
		check = "[✓] "
	}

	var line string
	if u.name == updateAllSentinel {
		line = cursor + check + "Update all"
	} else if u.locked != "" {
		line = cursor + check + u.name + "  " + u.constraint + "  (locked: " + u.locked + ")"
	} else {
		line = cursor + check + u.name + "  " + u.constraint
	}

	if d.selected[index] {
		fmt.Fprintln(w, selectedItem.Render(line))
	} else if index == m.Index() {
		fmt.Fprintln(w, normalItem.Render(line))
	} else {
		fmt.Fprintln(w, dimItem.Render(line))
	}
}

// updateModel is a bubbletea model for the interactive update skill list.
type updateModel struct {
	list     list.Model
	sel      map[int]bool
	selected []string // dep names chosen by the user
	allDeps  []string // all semver-constrained dep names (for "Update all")
	quitting bool
}

func newUpdateModel(skills []updateSkillItem, allDeps []string) updateModel {
	// Prepend the "Update all" sentinel as index 0.
	items := make([]list.Item, 0, len(skills)+1)
	items = append(items, updateSkillItem{name: updateAllSentinel})
	for _, s := range skills {
		items = append(items, s)
	}

	height := len(items) + 2
	if height > 22 {
		height = 22
	}

	sel := map[int]bool{}
	l := list.New(items, updateMultiSelectDelegate{selected: sel}, 80, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return updateModel{list: l, sel: sel, allDeps: allDeps}
}

func (m updateModel) Init() tea.Cmd { return nil }

func (m updateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		available := msg.Height - listReservedRows
		if available < 2 {
			available = 2
		}
		if available < m.list.Height() {
			m.list.SetHeight(available)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeySpace:
			idx := m.list.Index()
			m.sel[idx] = !m.sel[idx]
			m.list.SetDelegate(updateMultiSelectDelegate{selected: m.sel})
			return m, nil
		case tea.KeyEnter:
			if m.sel[0] {
				// "Update all" sentinel selected — treat all deps as selected.
				m.selected = m.allDeps
			} else {
				items := m.list.Items()
				for i, item := range items {
					if i == 0 {
						continue // skip sentinel
					}
					if m.sel[i] {
						m.selected = append(m.selected, item.(updateSkillItem).name)
					}
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

func (m updateModel) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select skills to update") + "\n")
	b.WriteString(m.list.View() + "\n")
	b.WriteString(hintStyle.Render("↑↓ navigate  space to toggle  enter to confirm  esc to cancel") + "\n")
	return b.String()
}

// runUpdateTUI runs the bubbletea update list and returns the selected dep names.
func runUpdateTUI(skills []updateSkillItem, allDeps []string) ([]string, error) {
	m := newUpdateModel(skills, allDeps)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(updateModel).selected, nil
}
