package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDeck(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "plain flashcards tag",
			lines: []string{"some text", "#flashcards"},
			want:  "default",
		},
		{
			name:  "flashcards with topic",
			lines: []string{"#flashcards/math"},
			want:  "math",
		},
		{
			name:  "nested flashcards tag",
			lines: []string{"#flashcards/cs/algorithms"},
			want:  "cs.algorithms",
		},
		{
			name:  "no flashcards tag",
			lines: []string{"some text", "more text"},
			want:  "",
		},
		{
			name:  "tag mid-line with other words",
			lines: []string{"tags: #flashcards/history #review"},
			want:  "history",
		},
		{
			name:  "deeply nested tag",
			lines: []string{"#flashcards/a/b/c"},
			want:  "a.b.c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDeck(tt.lines)
			if got != tt.want {
				t.Errorf("findDeck() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractCards(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int // expected card count
		qFirst string // expected question of first card
		aFirst string // expected answer of first card
	}{
		{
			name: "single card",
			lines: []string{
				"What is 2+2?",
				"?",
				"4",
			},
			want:   1,
			qFirst: "What is 2+2?",
			aFirst: "4",
		},
		{
			name: "multiple cards",
			lines: []string{
				"Question one",
				"?",
				"Answer one",
				"",
				"Question two",
				"?",
				"Answer two",
			},
			want:   2,
			qFirst: "Question one",
			aFirst: "Answer one",
		},
		{
			name: "multi-line question",
			lines: []string{
				"Line one of question",
				"Line two of question",
				"?",
				"The answer",
			},
			want:   1,
			qFirst: "Line one of question\nLine two of question",
			aFirst: "The answer",
		},
		{
			name: "no separator",
			lines: []string{
				"Just some text",
				"No question mark separator",
			},
			want: 0,
		},
		{
			name: "blank lines between cards",
			lines: []string{
				"Q1",
				"?",
				"A1",
				"",
				"",
				"Q2",
				"?",
				"A2",
			},
			want:   2,
			qFirst: "Q1",
			aFirst: "A1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := extractCards(tt.lines, "test-deck", "test.md")
			if len(cards) != tt.want {
				t.Fatalf("got %d cards, want %d", len(cards), tt.want)
			}
			if tt.want > 0 {
				if cards[0].Question != tt.qFirst {
					t.Errorf("first question = %q, want %q", cards[0].Question, tt.qFirst)
				}
				if cards[0].Answer != tt.aFirst {
					t.Errorf("first answer = %q, want %q", cards[0].Answer, tt.aFirst)
				}
				if cards[0].DeckName != "test-deck" {
					t.Errorf("deck = %q, want %q", cards[0].DeckName, "test-deck")
				}
			}
		})
	}
}

func TestParseDirectory(t *testing.T) {
	dir := t.TempDir()

	// File with flashcards
	content1 := `# Notes
Some intro text

What is Go?
?
A programming language

What is Rust?
?
A systems programming language

#flashcards/programming
`
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}

	// File without flashcards tag
	content2 := `# Random notes
Nothing to see here
`
	if err := os.WriteFile(filepath.Join(dir, "random.md"), []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}

	// Non-markdown file (should be skipped)
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Hidden directory (should be skipped)
	hiddenDir := filepath.Join(dir, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	content3 := `What?
?
Hidden
#flashcards
`
	if err := os.WriteFile(filepath.Join(hiddenDir, "secret.md"), []byte(content3), 0644); err != nil {
		t.Fatal(err)
	}

	cards, err := ParseDirectory(dir)
	if err != nil {
		t.Fatalf("ParseDirectory() error: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("got %d cards, want 2", len(cards))
	}

	for _, c := range cards {
		if c.DeckName != "programming" {
			t.Errorf("deck = %q, want %q", c.DeckName, "programming")
		}
	}
}
