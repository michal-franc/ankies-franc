package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michal-franc/ankies-franc/config"
	"github.com/michal-franc/ankies-franc/parser"
	"github.com/michal-franc/ankies-franc/storage"
	"github.com/michal-franc/ankies-franc/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()
	cmd := os.Args[1]
	rest := os.Args[2:]

	// Parse args: first non-flag arg is the path, rest are flags
	pathArg := ""
	dueFormat := "plain"
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--json":
			dueFormat = "json"
		case "--by-deck":
			dueFormat = "by-deck"
		case "--format":
			if i+1 < len(rest) {
				i++
				dueFormat = rest[i]
			}
		default:
			if pathArg == "" {
				pathArg = rest[i]
			}
		}
	}

	notesPath := cfg.ResolvePath(pathArg)

	if notesPath == "" {
		fmt.Fprintln(os.Stderr, "No path provided and no notes_path in config.")
		fmt.Fprintf(os.Stderr, "Set it in %s or pass a path argument.\n", config.DefaultConfigPath())
		os.Exit(1)
	}

	switch cmd {
	case "review":
		runReview(notesPath, cfg)
	case "due":
		runDue(notesPath, dueFormat, cfg)
	case "list":
		runList(notesPath, cfg)
	case "config":
		runConfig(notesPath, cfg)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: ankies-franc <command> [path] [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  review  Interactive TUI review of due cards")
	fmt.Fprintln(os.Stderr, "  due     Print count of due cards (for polybar)")
	fmt.Fprintln(os.Stderr, "  list    List decks and card counts")
	fmt.Fprintln(os.Stderr, "  config  Configure deck ignore list")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Due flags:")
	fmt.Fprintln(os.Stderr, "  --json            JSON output with full stats")
	fmt.Fprintln(os.Stderr, "  --format polybar  One-liner with new/overdue breakdown")
	fmt.Fprintln(os.Stderr, "  --by-deck         Due counts per deck")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Path is optional if notes_path is set in %s\n", config.DefaultConfigPath())
}

func filterIgnored(cards []parser.Card, cfg config.Config) []parser.Card {
	var filtered []parser.Card
	for _, c := range cards {
		if !cfg.IsDeckIgnored(c.DeckName) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func runReview(path string, cfg config.Config) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}
	cards = filterIgnored(cards, cfg)

	store, err := storage.Load(storage.DefaultPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	if len(cards) == 0 {
		fmt.Println("No flashcards found.")
		return
	}

	model := tui.New(cards, store)
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

func runDue(path string, format string, cfg config.Config) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}
	cards = filterIgnored(cards, cfg)

	store, err := storage.Load(storage.DefaultPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Compute stats
	var questions []string
	type deckStats struct {
		Due   int `json:"due"`
		New   int `json:"new"`
		Total int `json:"total"`
	}
	decks := make(map[string]*deckStats)

	totalDue, totalNew, totalOverdue := 0, 0, 0
	for _, c := range cards {
		questions = append(questions, c.Question)

		ds, ok := decks[c.DeckName]
		if !ok {
			ds = &deckStats{}
			decks[c.DeckName] = ds
		}
		ds.Total++

		if store.IsDue(c.Question) {
			totalDue++
			ds.Due++
		}
		if store.IsNew(c.Question) {
			totalNew++
			ds.New++
		}
		if store.IsOverdue(c.Question) {
			totalOverdue++
		}
	}

	reviewedToday := store.ReviewedToday(questions)
	streak := store.Streak()

	switch format {
	case "json":
		output := struct {
			Due           int                   `json:"due"`
			New           int                   `json:"new"`
			Overdue       int                   `json:"overdue"`
			ReviewedToday int                   `json:"reviewed_today"`
			Streak        int                   `json:"streak"`
			Decks         map[string]*deckStats  `json:"decks"`
		}{
			Due:           totalDue,
			New:           totalNew,
			Overdue:       totalOverdue,
			ReviewedToday: reviewedToday,
			Streak:        streak,
			Decks:         decks,
		}
		data, _ := json.Marshal(output)
		fmt.Println(string(data))

	case "polybar":
		parts := []string{}
		if totalNew > 0 {
			parts = append(parts, fmt.Sprintf("%d new", totalNew))
		}
		if totalOverdue > 0 {
			parts = append(parts, fmt.Sprintf("%d overdue", totalOverdue))
		}
		if len(parts) > 0 {
			fmt.Printf("%d (%s)\n", totalDue, strings.Join(parts, ", "))
		} else {
			fmt.Println(totalDue)
		}

	case "by-deck":
		var names []string
		for name := range decks {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := []string{}
		for _, name := range names {
			if decks[name].Due > 0 {
				parts = append(parts, fmt.Sprintf("%s: %d", name, decks[name].Due))
			}
		}
		fmt.Println(strings.Join(parts, "  "))

	default:
		fmt.Println(totalDue)
	}

	if totalDue == 0 {
		os.Exit(1)
	}
}

func runConfig(path string, cfg config.Config) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}

	// Build deck list from ALL cards (no ignore filtering)
	deckCards := make(map[string]int)
	for _, c := range cards {
		deckCards[c.DeckName]++
	}

	var names []string
	for name := range deckCards {
		names = append(names, name)
	}
	sort.Strings(names)

	var entries []tui.ConfigEntry
	for _, name := range names {
		entries = append(entries, tui.ConfigEntry{
			Name:    name,
			Cards:   deckCards[name],
			Ignored: cfg.IsDeckIgnored(name),
		})
	}

	model := tui.NewConfigModel(entries)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	result := finalModel.(tui.ConfigModel).Result()
	if result.Saved {
		cfg.IgnoreDecks = result.IgnoreDecks
		if err := cfg.Save(config.DefaultConfigPath()); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
	}
}

func runList(path string, cfg config.Config) {
	cards, err := parser.ParseDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cards: %v\n", err)
		os.Exit(1)
	}
	cards = filterIgnored(cards, cfg)

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
