// Package input provides a shell script interface for the text input bubble.
// https://github.com/charmbracelet/bubbles/tree/master/textinput
//
// It can be used to prompt the user for some input. The text the user entered
// will be sent to stdout.
//
// $ gum input --placeholder "What's your favorite gum?" > answer.text
package input

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/gum/timeout"
	"github.com/charmbracelet/lipgloss"
)

type keymap textinput.KeyMap

func defaultKeymap() keymap {
	k := textinput.DefaultKeyMap
	return keymap(k)
}

// FullHelp implements help.KeyMap.
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// ShortHelp implements help.KeyMap.
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
	}
}

type model struct {
	autoWidth   bool
	header      string
	headerStyle lipgloss.Style
	textinput   textinput.Model
	quitting    bool
	timedOut    bool
	aborted     bool
	timeout     time.Duration
	hasTimeout  bool
	showHelp    bool
	help        help.Model
	keymap      keymap
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		timeout.Init(m.timeout, nil),
	)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.header != "" {
		header := m.headerStyle.Render(m.header)
		return lipgloss.JoinVertical(lipgloss.Left, header, m.textinput.View())
	}

	if !m.showHelp {
		return m.textinput.View()
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.textinput.View(),
		"",
		m.help.View(m.keymap),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeout.TickTimeoutMsg:
		if msg.TimeoutValue <= 0 {
			m.quitting = true
			m.timedOut = true
			return m, tea.Quit
		}
		m.timeout = msg.TimeoutValue
		return m, timeout.Tick(msg.TimeoutValue, msg.Data)
	case tea.WindowSizeMsg:
		if m.autoWidth {
			m.textinput.Width = msg.Width - lipgloss.Width(m.textinput.Prompt) - 1
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			m.aborted = true
			return m, tea.Quit
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}
