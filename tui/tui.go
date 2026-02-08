package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michal-franc/ankies-franc/parser"
	"github.com/michal-franc/ankies-franc/storage"
)

type state int

const (
	showingQuestion state = iota
	showingAnswer
	done
)

type Model struct {
	cards    []parser.Card
	store    *storage.Store
	current  int
	state    state
	total    int
	reviewed int
	quitting bool
}

func New(cards []parser.Card, store *storage.Store) Model {
	return Model{
		cards: cards,
		store: store,
		total: len(cards),
		state: showingQuestion,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
)

func (m Model) View() string {
	if m.quitting {
		return fmt.Sprintf("Reviewed %d cards. State saved.\n", m.reviewed)
	}

	if m.state == done {
		return doneStyle.Render(fmt.Sprintf("Done for today! Reviewed %d cards.\n", m.reviewed))
	}

	card := m.cards[m.current]
	var b strings.Builder

	// Header: deck + progress
	header := fmt.Sprintf("%s  %s",
		deckStyle.Render(card.DeckName),
		progressStyle.Render(fmt.Sprintf("%d/%d", m.current+1, m.total)),
	)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	// Question
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
