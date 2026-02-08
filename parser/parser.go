package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Card struct {
	DeckName   string
	Question   string
	Answer     string
	SourceFile string
}

func ParseDirectory(root string) ([]Card, error) {
	var cards []Card

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}
		if info.IsDir() {
			// skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		fileCards, err := parseFile(path)
		if err != nil {
			return nil // skip unparseable files
		}
		cards = append(cards, fileCards...)
		return nil
	})

	return cards, err
}

func parseFile(path string) ([]Card, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Find the #flashcards tag and extract deck name
	deck := findDeck(lines)
	if deck == "" {
		return nil, nil // no flashcards tag found
	}

	return extractCards(lines, deck, path), nil
}

func findDeck(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// look for #flashcards tag anywhere in the line
		for _, word := range strings.Fields(trimmed) {
			if word == "#flashcards" {
				return "default"
			}
			if strings.HasPrefix(word, "#flashcards/") {
				// extract deck name: everything after #flashcards/
				deck := strings.TrimPrefix(word, "#flashcards/")
				// replace / with . for nested decks
				deck = strings.ReplaceAll(deck, "/", ".")
				return deck
			}
		}
	}
	return ""
}

func extractCards(lines []string, deck, sourceFile string) []Card {
	// Strategy: find all ? separators, then for each one:
	// - Question = lines before the ?, going back to the previous card boundary
	// - Answer = lines after the ?, going forward to the next card boundary
	//
	// Card boundaries are detected by looking at the region between two ? separators:
	// the LAST group of contiguous non-blank lines before a ? is that card's question;
	// everything before that group (after the previous ?) is the previous card's answer.

	var sepIndices []int
	for i, line := range lines {
		if strings.TrimSpace(line) == "?" {
			sepIndices = append(sepIndices, i)
		}
	}

	if len(sepIndices) == 0 {
		return nil
	}

	// For each ? at index sepIndices[i], find the question start.
	// The question is the last block of non-blank, non-tag content lines
	// before the ?, bounded by either: a blank line, a #review-flashcard tag,
	// or the previous ? separator.
	type cardRange struct {
		qStart, qEnd int // question line range [qStart, qEnd)
		aStart, aEnd int // answer line range [aStart, aEnd)
	}

	var ranges []cardRange

	for si, sepIdx := range sepIndices {
		// Determine the region where the question can be found
		regionStart := 0
		if si > 0 {
			regionStart = sepIndices[si-1] + 1
		}

		// Find question: walk backward from sepIdx-1, skip blank lines,
		// then collect contiguous non-blank lines until we hit a blank line,
		// #review-flashcard, flashcards tag, or regionStart
		qEnd := sepIdx // exclusive
		qStart := sepIdx

		// Skip trailing blanks before ?
		for qEnd > regionStart && strings.TrimSpace(lines[qEnd-1]) == "" {
			qEnd--
		}

		// Now walk backwards to find the start of the question block
		qStart = qEnd
		for qStart > regionStart {
			trimmed := strings.TrimSpace(lines[qStart-1])
			if trimmed == "" || trimmed == "#review-flashcard" || containsFlashcardsTag(trimmed) {
				break
			}
			qStart--
		}

		// Answer region: from sepIdx+1 to next card's question start (or EOF)
		aStart := sepIdx + 1
		aEnd := len(lines)

		// If there's a next separator, the answer ends where that card's question starts
		// We'll fix this up in a second pass after computing all question ranges
		ranges = append(ranges, cardRange{qStart, qEnd, aStart, aEnd})
	}

	// Fix up answer end: each answer ends where the next card's question starts
	for i := 0; i < len(ranges)-1; i++ {
		// The next card's question might have #review-flashcard before it
		// Answer ends at the first #review-flashcard or the question start, whichever is earlier
		nextQStart := ranges[i+1].qStart
		// Check if there's a #review-flashcard between our answer start and next question
		aEnd := nextQStart
		for j := ranges[i].aStart; j < nextQStart; j++ {
			if strings.TrimSpace(lines[j]) == "#review-flashcard" {
				aEnd = j
				break
			}
		}
		ranges[i].aEnd = aEnd
	}

	var cards []Card
	for _, r := range ranges {
		var qLines, aLines []string
		for i := r.qStart; i < r.qEnd; i++ {
			qLines = append(qLines, lines[i])
		}
		for i := r.aStart; i < r.aEnd; i++ {
			aLines = append(aLines, lines[i])
		}

		q := strings.TrimSpace(strings.Join(qLines, "\n"))
		a := strings.TrimSpace(strings.Join(aLines, "\n"))
		if q != "" {
			cards = append(cards, Card{
				DeckName:   deck,
				Question:   q,
				Answer:     a,
				SourceFile: sourceFile,
			})
		}
	}

	return cards
}

func containsFlashcardsTag(line string) bool {
	for _, word := range strings.Fields(line) {
		if word == "#flashcards" || strings.HasPrefix(word, "#flashcards/") {
			return true
		}
	}
	return false
}
