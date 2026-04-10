package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// removeSkillItem is a single entry in the interactive remove list.
type removeSkillItem struct {
	name    string
	version string
}

func (r removeSkillItem) FilterValue() string { return r.name }

// removeMultiSelectDelegate renders each installed skill as a single line with a checkbox:
//
//	> [✓] github.com/owner/skill  ^1.2.0
type removeMultiSelectDelegate struct {
	selected map[int]bool
}

func (d removeMultiSelectDelegate) Height() int                              { return 1 }
func (d removeMultiSelectDelegate) Spacing() int                             { return 0 }
func (d removeMultiSelectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d removeMultiSelectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	r := item.(removeSkillItem)

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	check := "[ ] "
	if d.selected[index] {
		check = "[✓] "
	}

	line := cursor + check + r.name + "  " + r.version

	if d.selected[index] {
		fmt.Fprintln(w, selectedItem.Render(line))
	} else if index == m.Index() {
		fmt.Fprintln(w, normalItem.Render(line))
	} else {
		fmt.Fprintln(w, dimItem.Render(line))
	}
}

// removeModel is a bubbletea model for the interactive remove skill list.
type removeModel struct {
	list     list.Model
	sel      map[int]bool
	selected []string
	quitting bool
}

func newRemoveModel(skills []removeSkillItem) removeModel {
	items := make([]list.Item, len(skills))
	for i, s := range skills {
		items[i] = s
	}

	height := len(skills)
	if height > 20 {
		height = 20
	}

	sel := map[int]bool{}
	l := list.New(items, removeMultiSelectDelegate{selected: sel}, 80, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return removeModel{list: l, sel: sel}
}

func (m removeModel) Init() tea.Cmd { return nil }

func (m removeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeySpace:
			idx := m.list.Index()
			m.sel[idx] = !m.sel[idx]
			m.list.SetDelegate(removeMultiSelectDelegate{selected: m.sel})
			return m, nil
		case tea.KeyEnter:
			items := m.list.Items()
			for i, item := range items {
				if m.sel[i] {
					m.selected = append(m.selected, item.(removeSkillItem).name)
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

func (m removeModel) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select skills to remove") + "\n")
	b.WriteString(m.list.View() + "\n")
	b.WriteString(hintStyle.Render("↑↓ navigate  space to toggle  enter to confirm  esc to cancel") + "\n")
	return b.String()
}

// runRemoveTUI runs the bubbletea remove list and returns the selected skill names.
func runRemoveTUI(skills []removeSkillItem) ([]string, error) {
	m := newRemoveModel(skills)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(removeModel).selected, nil
}
