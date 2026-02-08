package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfigEntry struct {
	Name    string
	Cards   int
	Ignored bool
}

type ConfigResult struct {
	IgnoreDecks []string
	Saved       bool
}

type ConfigModel struct {
	entries []ConfigEntry
	cursor  int
	saved   bool
}

func NewConfigModel(entries []ConfigEntry) ConfigModel {
	return ConfigModel{
		entries: entries,
	}
}

func (m ConfigModel) Result() ConfigResult {
	var ignored []string
	for _, e := range m.entries {
		if e.Ignored {
			ignored = append(ignored, e.Name)
		}
	}
	return ConfigResult{
		IgnoreDecks: ignored,
		Saved:       m.saved,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}

		case " ":
			m.entries[m.cursor].Ignored = !m.entries[m.cursor].Ignored

		case "a":
			allIgnored := true
			for _, e := range m.entries {
				if !e.Ignored {
					allIgnored = false
					break
				}
			}
			for i := range m.entries {
				m.entries[i].Ignored = !allIgnored
			}

		case "enter":
			m.saved = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConfigModel) View() string {
	if m.saved {
		return "Config saved.\n"
	}

	var b strings.Builder

	b.WriteString(pickerTitleStyle.Render("Configure deck ignore list"))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	for i, e := range m.entries {
		cursor := "  "
		if i == m.cursor {
			cursor = pickerCursorStyle.Render("> ")
		}

		status := pickerSelectedStyle.Render("[active] ")
		if e.Ignored {
			status = pickerUnselectedStyle.Render("[ignored]")
		}

		name := e.Name
		if i == m.cursor {
			name = pickerCursorStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s %-30s %3d cards\n",
			cursor, status, name, e.Cards))
	}

	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("[space] toggle  [a] toggle all  [enter] save & quit  [q] quit"))
	b.WriteString("\n")

	return b.String()
}
