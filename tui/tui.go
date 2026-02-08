package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michal-franc/ankies-franc/parser"
	"github.com/michal-franc/ankies-franc/storage"
)

type state int

const (
	pickingDecks state = iota
	showingQuestion
	showingAnswer
	done
)

type deckInfo struct {
	name     string
	due      int
	total    int
	selected bool
}

type Model struct {
	allCards []parser.Card
	cards    []parser.Card
	store    *storage.Store
	current  int
	state    state
	total    int
	reviewed int
	quitting bool

	// deck picker
	decks  []deckInfo
	cursor int
}

func New(cards []parser.Card, store *storage.Store) Model {
	// Build deck info
	deckMap := make(map[string]*deckInfo)
	for _, c := range cards {
		d, ok := deckMap[c.DeckName]
		if !ok {
			d = &deckInfo{name: c.DeckName, selected: true}
			deckMap[c.DeckName] = d
		}
		d.total++
		if store.IsDue(c.Question) {
			d.due++
		}
	}

	var decks []deckInfo
	for _, d := range deckMap {
		decks = append(decks, *d)
	}
	sort.Slice(decks, func(i, j int) bool {
		return decks[i].name < decks[j].name
	})

	return Model{
		allCards: cards,
		store:    store,
		state:    pickingDecks,
		decks:    decks,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case pickingDecks:
			return m.updateDeckPicker(msg)
		default:
			return m.updateReview(msg)
		}
	}
	return m, nil
}

func (m Model) updateDeckPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.decks)-1 {
			m.cursor++
		}

	case " ":
		m.decks[m.cursor].selected = !m.decks[m.cursor].selected

	case "a":
		allSelected := true
		for _, d := range m.decks {
			if !d.selected {
				allSelected = false
				break
			}
		}
		for i := range m.decks {
			m.decks[i].selected = !allSelected
		}

	case "enter":
		m = m.startReview()
	}

	return m, nil
}

func (m Model) startReview() Model {
	// Build selected deck set
	selected := make(map[string]bool)
	for _, d := range m.decks {
		if d.selected {
			selected[d.name] = true
		}
	}

	// Filter to due cards in selected decks
	var dueCards []parser.Card
	for _, c := range m.allCards {
		if selected[c.DeckName] && m.store.IsDue(c.Question) {
			dueCards = append(dueCards, c)
		}
	}

	m.cards = dueCards
	m.total = len(dueCards)
	m.current = 0

	if len(dueCards) == 0 {
		m.state = done
	} else {
		m.state = showingQuestion
	}

	return m
}

func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case " ":
		if m.state == showingQuestion {
			m.state = showingAnswer
		}

	case "1", "h":
		if m.state == showingAnswer {
			m.store.Rate(m.cards[m.current].Question, storage.Hard)
			m.reviewed++
			m = m.advance()
		}

	case "2", "g":
		if m.state == showingAnswer {
			m.store.Rate(m.cards[m.current].Question, storage.Good)
			m.reviewed++
			m = m.advance()
		}

	case "3", "e":
		if m.state == showingAnswer {
			m.store.Rate(m.cards[m.current].Question, storage.Easy)
			m.reviewed++
			m = m.advance()
		}
	}

	return m, nil
}

func (m Model) advance() Model {
	m.current++
	if m.current >= len(m.cards) {
		m.state = done
	} else {
		m.state = showingQuestion
	}
	return m
}

var (
	deckStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	questionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	answerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	ratingHardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	ratingGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	ratingEasyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	pickerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15"))

	pickerSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))

	pickerUnselectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	pickerCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("14"))

	pickerDueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	pickerNoDueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

func (m Model) View() string {
	if m.quitting {
		return fmt.Sprintf("Reviewed %d cards. State saved.\n", m.reviewed)
	}

	switch m.state {
	case pickingDecks:
		return m.viewDeckPicker()
	case done:
		return doneStyle.Render(fmt.Sprintf("Done for today! Reviewed %d cards.\n", m.reviewed))
	default:
		return m.viewCard()
	}
}

func (m Model) viewDeckPicker() string {
	var b strings.Builder

	b.WriteString(pickerTitleStyle.Render("Select decks to review"))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	totalDue := 0
	for i, d := range m.decks {
		cursor := "  "
		if i == m.cursor {
			cursor = pickerCursorStyle.Render("> ")
		}

		check := pickerUnselectedStyle.Render("[ ]")
		if d.selected {
			check = pickerSelectedStyle.Render("[x]")
		}

		dueStr := pickerNoDueStyle.Render("0 due")
		if d.due > 0 {
			dueStr = pickerDueStyle.Render(fmt.Sprintf("%d due", d.due))
			if d.selected {
				totalDue += d.due
			}
		}

		name := d.name
		if i == m.cursor {
			name = pickerCursorStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s %-30s %3d cards  %s\n",
			cursor, check, name, d.total, dueStr))
	}

	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Selected: %s\n", pickerDueStyle.Render(fmt.Sprintf("%d cards due", totalDue))))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("[space] toggle  [a] toggle all  [enter] start  [q] quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewCard() string {
	card := m.cards[m.current]
	var b strings.Builder

	header := fmt.Sprintf("%s  %s",
		deckStyle.Render(card.DeckName),
		progressStyle.Render(fmt.Sprintf("%d/%d", m.current+1, m.total)),
	)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	b.WriteString(questionStyle.Render(card.Question))
	b.WriteString("\n")

	if m.state == showingAnswer {
		b.WriteString("\n")
		b.WriteString(separatorStyle.Render("───"))
		b.WriteString("\n\n")
		b.WriteString(answerStyle.Render(card.Answer))
		b.WriteString("\n\n")
		b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s  %s  %s",
			ratingHardStyle.Render("[1/h] Hard"),
			ratingGoodStyle.Render("[2/g] Good"),
			ratingEasyStyle.Render("[3/e] Easy"),
		))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("[space] flip  [q] quit"))
		b.WriteString("\n")
	}

	return b.String()
}
