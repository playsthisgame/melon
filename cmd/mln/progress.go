package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type depFetchedMsg struct {
	index int
	name  string
	total int
	err   error
}

type fetchDoneMsg struct {
	count int
}

type installProgressModel struct {
	bar     progress.Model
	label   string
	total   int
	done    int
	err     error
	quitting bool
}

var labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

func newInstallProgressModel(total int) installProgressModel {
	bar := progress.New(progress.WithDefaultGradient())
	return installProgressModel{
		bar:   bar,
		total: total,
	}
}

func (m installProgressModel) Init() tea.Cmd {
	return nil
}

func (m installProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case depFetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.quitting = true
			return m, tea.Quit
		}
		m.done = msg.index + 1
		m.label = fmt.Sprintf("Fetching %s…", msg.name)
		var pct float64
		if m.total > 0 {
			pct = float64(m.done) / float64(m.total)
		}
		cmd := m.bar.SetPercent(pct)
		return m, cmd

	case fetchDoneMsg:
		m.quitting = true
		cmd := m.bar.SetPercent(1.0)
		return m, tea.Sequence(cmd, tea.Quit)

	case progress.FrameMsg:
		barModel, cmd := m.bar.Update(msg)
		m.bar = barModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m installProgressModel) View() string {
	if m.quitting {
		return ""
	}
	label := labelStyle.Render(m.label)
	return "\n" + m.bar.View() + "\n" + label + "\n"
}
