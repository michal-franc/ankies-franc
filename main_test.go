package main

import (
	"testing"

	"github.com/michal-franc/ankies-franc/config"
	"github.com/michal-franc/ankies-franc/parser"
)

func TestFilterIgnored(t *testing.T) {
	tests := []struct {
		name        string
		cards       []parser.Card
		ignoreDecks []string
		wantCount   int
		wantDecks   []string
	}{
		{
			name: "filters matching decks",
			cards: []parser.Card{
				{DeckName: "golang", Question: "Q1"},
				{DeckName: "leetcode", Question: "Q2"},
				{DeckName: "math", Question: "Q3"},
			},
			ignoreDecks: []string{"leetcode"},
			wantCount:   2,
			wantDecks:   []string{"golang", "math"},
		},
		{
			name: "filters prefix matches",
			cards: []parser.Card{
				{DeckName: "leetcode.dp", Question: "Q1"},
				{DeckName: "leetcode.arrays", Question: "Q2"},
				{DeckName: "math", Question: "Q3"},
			},
			ignoreDecks: []string{"leetcode"},
			wantCount:   1,
			wantDecks:   []string{"math"},
		},
		{
			name: "empty ignore list keeps all",
			cards: []parser.Card{
				{DeckName: "golang", Question: "Q1"},
				{DeckName: "math", Question: "Q2"},
			},
			ignoreDecks: []string{},
			wantCount:   2,
			wantDecks:   []string{"golang", "math"},
		},
		{
			name:        "no cards returns empty",
			cards:       []parser.Card{},
			ignoreDecks: []string{"leetcode"},
			wantCount:   0,
		},
		{
			name: "multiple ignore patterns",
			cards: []parser.Card{
				{DeckName: "golang", Question: "Q1"},
				{DeckName: "leetcode", Question: "Q2"},
				{DeckName: "python", Question: "Q3"},
				{DeckName: "math", Question: "Q4"},
			},
			ignoreDecks: []string{"leetcode", "python"},
			wantCount:   2,
			wantDecks:   []string{"golang", "math"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{IgnoreDecks: tt.ignoreDecks}
			got := filterIgnored(tt.cards, cfg)

			if len(got) != tt.wantCount {
				t.Fatalf("got %d cards, want %d", len(got), tt.wantCount)
			}

			for i, deck := range tt.wantDecks {
				if got[i].DeckName != deck {
					t.Errorf("card[%d].DeckName = %q, want %q", i, got[i].DeckName, deck)
				}
			}
		})
	}
}
