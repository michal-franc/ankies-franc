package main

import (
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michal-franc/ankies-franc/parser"
	"github.com/michal-franc/ankies-franc/storage"
	"github.com/michal-franc/ankies-franc/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "review":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ankies-franc review <path>")
			os.Exit(1)
		}
		runReview(os.Args[2])
	case "due":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ankies-franc due <path>")
			os.Exit(1)
		}
		runDue(os.Args[2])
	case "list":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ankies-franc list <path>")
			os.Exit(1)
		}
		runList(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: ankies-franc <command> <path>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  review  Interactive TUI review of due cards")
	fmt.Fprintln(os.Stderr, "  due     Print count of due cards (for polybar)")
	fmt.Fprintln(os.Stderr, "  list    List decks and card counts")
}

func runReview(path string) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.Load(storage.DefaultPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Filter to only due cards
	var dueCards []parser.Card
	for _, c := range cards {
		if store.IsDue(c.Question) {
			dueCards = append(dueCards, c)
		}
	}

	if len(dueCards) == 0 {
		fmt.Println("No cards due for review. Come back later!")
		return
	}

	model := tui.New(dueCards, store)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Save state after review
	_ = finalModel // state was already mutated via pointer
	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
		os.Exit(1)
	}
}

func runDue(path string) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.Load(storage.DefaultPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	var questions []string
	for _, c := range cards {
		questions = append(questions, c.Question)
	}

	count := store.DueCount(questions)
	fmt.Println(count)

	if count == 0 {
		os.Exit(1)
	}
}

func runList(path string) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.Load(storage.DefaultPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Group by deck
	decks := make(map[string]struct {
		total int
		due   int
	})

	for _, c := range cards {
		d := decks[c.DeckName]
		d.total++
		if store.IsDue(c.Question) {
			d.due++
		}
		decks[c.DeckName] = d
	}

	// Sort deck names
	var names []string
	for name := range decks {
		names = append(names, name)
	}
	sort.Strings(names)

	totalCards := 0
	totalDue := 0
	for _, name := range names {
		d := decks[name]
		fmt.Printf("%-30s %3d cards  (%d due)\n", name, d.total, d.due)
		totalCards += d.total
		totalDue += d.due
	}
	fmt.Printf("%-30s %3d cards  (%d due)\n", "TOTAL", totalCards, totalDue)
}
