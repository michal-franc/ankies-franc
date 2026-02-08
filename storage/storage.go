package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CardState struct {
	NextReview   time.Time `json:"next_review"`
	Interval     int       `json:"interval"` // days
	EaseFactor   float64   `json:"ease_factor"`
	LastReviewed time.Time `json:"last_reviewed,omitempty"`
}

type Rating int

const (
	Hard Rating = iota
	Good
	Easy
)

type Store struct {
	Cards map[string]CardState `json:"cards"`
	path  string
}

func CardKey(question string) string {
	h := sha256.Sum256([]byte(question))
	return fmt.Sprintf("%x", h[:16])
}

func DefaultPath() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "ankies-franc", "state.json")
}

func Load(path string) (*Store, error) {
	s := &Store{
		Cards: make(map[string]CardState),
		path:  path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &s.Cards); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.Cards, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) GetState(question string) CardState {
	key := CardKey(question)
	state, ok := s.Cards[key]
	if !ok {
		return CardState{
			NextReview: time.Time{}, // zero time = due now
			Interval:   0,
			EaseFactor: 2.5,
		}
	}
	return state
}

func (s *Store) IsDue(question string) bool {
	state := s.GetState(question)
	return !time.Now().Before(state.NextReview)
}

func (s *Store) Rate(question string, rating Rating) {
	key := CardKey(question)
	state := s.GetState(question)

	if state.Interval == 0 {
		// New card
		state.Interval = 1
		state.EaseFactor = 2.5
	}

	switch rating {
	case Hard:
		// interval stays same, ease decreases
		state.EaseFactor -= 0.15
		if state.EaseFactor < 1.3 {
			state.EaseFactor = 1.3
		}
	case Good:
		// interval *= ease factor
		state.Interval = int(float64(state.Interval) * state.EaseFactor)
		if state.Interval < 1 {
			state.Interval = 1
		}
	case Easy:
		// interval *= ease * 1.3, ease increases
		state.Interval = int(float64(state.Interval) * state.EaseFactor * 1.3)
		if state.Interval < 1 {
			state.Interval = 1
		}
		state.EaseFactor += 0.15
	}

	now := time.Now()
	state.NextReview = now.Add(time.Duration(state.Interval) * 24 * time.Hour)
	state.LastReviewed = now
	s.Cards[key] = state
}

func (s *Store) DueCount(questions []string) int {
	count := 0
	for _, q := range questions {
		if s.IsDue(q) {
			count++
		}
	}
	return count
}

// IsNew returns true if the card has never been reviewed.
func (s *Store) IsNew(question string) bool {
	key := CardKey(question)
	_, ok := s.Cards[key]
	return !ok
}

// IsOverdue returns true if the card was due more than 1 day ago.
func (s *Store) IsOverdue(question string) bool {
	key := CardKey(question)
	state, ok := s.Cards[key]
	if !ok {
		return false // new cards aren't overdue
	}
	return time.Now().After(state.NextReview.Add(24 * time.Hour))
}

// ReviewedToday returns how many of the given questions were reviewed today.
func (s *Store) ReviewedToday(questions []string) int {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	count := 0
	for _, q := range questions {
		key := CardKey(q)
		state, ok := s.Cards[key]
		if ok && !state.LastReviewed.Before(startOfDay) {
			count++
		}
	}
	return count
}

// Streak returns the number of consecutive days (including today) with at least one review.
func (s *Store) Streak() int {
	if len(s.Cards) == 0 {
		return 0
	}

	// Collect unique review dates
	reviewDays := make(map[string]bool)
	for _, state := range s.Cards {
		if !state.LastReviewed.IsZero() {
			day := state.LastReviewed.Format("2006-01-02")
			reviewDays[day] = true
		}
	}

	if len(reviewDays) == 0 {
		return 0
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check if today has reviews; if not, start from yesterday
	streak := 0
	day := today
	if !reviewDays[day.Format("2006-01-02")] {
		// No review today, check if yesterday had one (streak from yesterday)
		day = day.AddDate(0, 0, -1)
		if !reviewDays[day.Format("2006-01-02")] {
			return 0
		}
	}

	for reviewDays[day.Format("2006-01-02")] {
		streak++
		day = day.AddDate(0, 0, -1)
	}

	return streak
}
