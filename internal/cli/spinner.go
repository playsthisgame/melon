package cli

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type spinnerModel struct {
	spinner  spinner.Model
	label    string
	done     bool
	err      error
	resultCh <-chan error
}

type spinnerDoneMsg struct{ err error }

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.waitForResult())
}

func (m spinnerModel) waitForResult() tea.Cmd {
	return func() tea.Msg {
		return spinnerDoneMsg{err: <-m.resultCh}
	}
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + m.label + "\n"
}

// withSpinner runs fn while displaying a spinner labeled with label.
// No-ops (runs fn directly) when stdout is not a TTY.
func withSpinner(label string, fn func() error) error {
	if !isTTY() {
		return fn()
	}

	resultCh := make(chan error, 1)
	go func() { resultCh <- fn() }()

	s := spinner.New()
	s.Spinner = spinner.Dot

	m := spinnerModel{
		spinner:  s,
		label:    label,
		resultCh: resultCh,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	return finalModel.(spinnerModel).err
}
