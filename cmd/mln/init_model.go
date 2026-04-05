package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/playsthisgame/melon/internal/agents"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	selectedItem = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	normalItem   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	dimItem      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// ── Step enum ─────────────────────────────────────────────────────────────────

type initStep int

const (
	stepName initStep = iota
	stepDescription
	stepAgents
	stepDone
)

// ── List item ─────────────────────────────────────────────────────────────────

type stringItem string

func (s stringItem) FilterValue() string { return string(s) }

// ── Multi-select delegate ─────────────────────────────────────────────────────

type multiSelectDelegate struct {
	selected map[int]bool
}

func (d multiSelectDelegate) Height() int                              { return 1 }
func (d multiSelectDelegate) Spacing() int                             { return 0 }
func (d multiSelectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d multiSelectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	label := string(item.(stringItem))
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}
	check := "[ ] "
	if d.selected[index] {
		check = "[✓] "
	}
	line := cursor + check + label
	if d.selected[index] {
		fmt.Fprintln(w, selectedItem.Render(line))
	} else if index == m.Index() {
		fmt.Fprintln(w, normalItem.Render(line))
	} else {
		fmt.Fprintln(w, dimItem.Render(line))
	}
}

// ── initResult ────────────────────────────────────────────────────────────────

type initResult struct {
	name        string
	description string
	agentNames  []string
}

// ── initModel ─────────────────────────────────────────────────────────────────

type initModel struct {
	step        initStep
	nameInput   textinput.Model
	descInput   textinput.Model
	agentList   list.Model
	agentSel    map[int]bool
	defaultName string
	result      initResult
	quitting    bool
}

func newInitModel(defaultName string) initModel {
	// Name input
	ni := textinput.New()
	ni.Placeholder = defaultName
	ni.Focus()

	// Description input
	di := textinput.New()
	di.Placeholder = "Short description (optional)"

	// Agent list — sorted alphabetically
	known := agents.KnownAgents()
	sort.Strings(known)
	agentItems := make([]list.Item, len(known))
	for i, a := range known {
		agentItems[i] = stringItem(a)
	}
	agentSel := map[int]bool{}
	al := list.New(agentItems, multiSelectDelegate{selected: agentSel}, 40, len(known)+2)
	al.SetShowTitle(false)
	al.SetShowStatusBar(false)
	al.SetFilteringEnabled(false)
	al.SetShowHelp(false)

	return initModel{
		step:        stepName,
		nameInput:   ni,
		descInput:   di,
		agentList:   al,
		agentSel:    agentSel,
		defaultName: defaultName,
	}
}

func (m initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.advance()

		case tea.KeySpace:
			if m.step == stepAgents {
				idx := m.agentList.Index()
				m.agentSel[idx] = !m.agentSel[idx]
				m.agentList.SetDelegate(multiSelectDelegate{selected: m.agentSel})
				return m, nil
			}
		}
	}

	return m.updateCurrentStep(msg)
}

func (m initModel) advance() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepName:
		val := strings.TrimSpace(m.nameInput.Value())
		if val == "" {
			val = m.defaultName
		}
		m.result.name = val
		m.step = stepDescription
		m.descInput.Focus()
		return m, textinput.Blink

	case stepDescription:
		m.result.description = strings.TrimSpace(m.descInput.Value())
		m.step = stepAgents
		return m, nil

	case stepAgents:
		known := agents.KnownAgents()
		sort.Strings(known)
		var selected []string
		for i, a := range known {
			if m.agentSel[i] {
				selected = append(selected, a)
			}
		}
		m.result.agentNames = selected
		m.step = stepDone
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m initModel) updateCurrentStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.step {
	case stepName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case stepDescription:
		m.descInput, cmd = m.descInput.Update(msg)
	case stepAgents:
		m.agentList, cmd = m.agentList.Update(msg)
	}
	return m, cmd
}

func (m initModel) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder

	switch m.step {
	case stepName:
		b.WriteString(titleStyle.Render("Project name") + "\n")
		b.WriteString(m.nameInput.View() + "\n")
		b.WriteString(hintStyle.Render("enter to confirm") + "\n")

	case stepDescription:
		b.WriteString(titleStyle.Render("Short description") + "\n")
		b.WriteString(m.descInput.View() + "\n")
		b.WriteString(hintStyle.Render("enter to confirm") + "\n")

	case stepAgents:
		b.WriteString(titleStyle.Render("AI tools") + "\n")
		b.WriteString(m.agentList.View() + "\n")
		b.WriteString(hintStyle.Render("↑↓ navigate  space to toggle  enter to confirm") + "\n")
	}

	return b.String()
}
