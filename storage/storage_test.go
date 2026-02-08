package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCardKey(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		key1 := CardKey("What is Go?")
		key2 := CardKey("What is Go?")
		if key1 != key2 {
			t.Errorf("same input produced different keys: %q vs %q", key1, key2)
		}
	})

	t.Run("different inputs produce different keys", func(t *testing.T) {
		key1 := CardKey("What is Go?")
		key2 := CardKey("What is Rust?")
		if key1 == key2 {
			t.Errorf("different inputs produced same key: %q", key1)
		}
	})
}

func TestLoadSave(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		store, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}

		// Rate a card to put data in the store
		store.Rate("test question", Good)

		if err := store.Save(); err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		// Reload
		store2, err := Load(path)
		if err != nil {
			t.Fatalf("Load() after save error: %v", err)
		}

		key := CardKey("test question")
		if _, ok := store2.Cards[key]; !ok {
			t.Fatal("saved card state not found after reload")
		}
	})

	t.Run("missing file returns empty store", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "nonexistent.json")

		store, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if len(store.Cards) != 0 {
			t.Errorf("expected empty store, got %d cards", len(store.Cards))
		}
	})
}

func TestGetState(t *testing.T) {
	t.Run("new card returns defaults", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		state := store.GetState("new question")

		if state.EaseFactor != 2.5 {
			t.Errorf("ease factor = %f, want 2.5", state.EaseFactor)
		}
		if !state.NextReview.IsZero() {
			t.Errorf("next review = %v, want zero time", state.NextReview)
		}
		if state.Interval != 0 {
			t.Errorf("interval = %d, want 0", state.Interval)
		}
	})

	t.Run("existing card returns stored state", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		key := CardKey("existing question")
		store.Cards[key] = CardState{
			EaseFactor: 3.0,
			Interval:   5,
			NextReview: time.Now().Add(24 * time.Hour),
		}

		state := store.GetState("existing question")
		if state.EaseFactor != 3.0 {
			t.Errorf("ease factor = %f, want 3.0", state.EaseFactor)
		}
		if state.Interval != 5 {
			t.Errorf("interval = %d, want 5", state.Interval)
		}
	})
}

func TestIsDue(t *testing.T) {
	t.Run("new card is due", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		if !store.IsDue("new question") {
			t.Error("new card should be due")
		}
	})

	t.Run("future card is not due", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		key := CardKey("future question")
		store.Cards[key] = CardState{
			NextReview: time.Now().Add(48 * time.Hour),
			EaseFactor: 2.5,
		}
		if store.IsDue("future question") {
			t.Error("future card should not be due")
		}
	})

	t.Run("past card is due", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		key := CardKey("past question")
		store.Cards[key] = CardState{
			NextReview: time.Now().Add(-24 * time.Hour),
			EaseFactor: 2.5,
		}
		if !store.IsDue("past question") {
			t.Error("past card should be due")
		}
	})
}

func TestRate(t *testing.T) {
	tests := []struct {
		name         string
		rating       Rating
		initInterval int
		initEase     float64
		wantMinInt   int
		wantEaseDiff float64 // expected change in ease factor
	}{
		{
			name:         "hard keeps interval and decreases ease",
			rating:       Hard,
			initInterval: 5,
			initEase:     2.5,
			wantMinInt:   5,
			wantEaseDiff: -0.15,
		},
		{
			name:         "good multiplies interval by ease",
			rating:       Good,
			initInterval: 5,
			initEase:     2.5,
			wantMinInt:   12, // int(5 * 2.5) = 12
			wantEaseDiff: 0,
		},
		{
			name:         "easy multiplies interval by ease times 1.3",
			rating:       Easy,
			initInterval: 5,
			initEase:     2.5,
			wantMinInt:   16, // int(5 * 2.5 * 1.3) = 16
			wantEaseDiff: 0.15,
		},
		{
			name:         "new card gets interval 1",
			rating:       Hard,
			initInterval: 0,
			initEase:     2.5,
			wantMinInt:   1,
			wantEaseDiff: -0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{Cards: make(map[string]CardState)}
			question := "rate test: " + tt.name

			if tt.initInterval > 0 {
				key := CardKey(question)
				store.Cards[key] = CardState{
					Interval:   tt.initInterval,
					EaseFactor: tt.initEase,
				}
			}

			store.Rate(question, tt.rating)

			key := CardKey(question)
			state := store.Cards[key]

			if state.Interval < tt.wantMinInt {
				t.Errorf("interval = %d, want >= %d", state.Interval, tt.wantMinInt)
			}

			expectedEase := tt.initEase + tt.wantEaseDiff
			if state.EaseFactor < expectedEase-0.01 || state.EaseFactor > expectedEase+0.01 {
				t.Errorf("ease factor = %f, want ~%f", state.EaseFactor, expectedEase)
			}

			if state.LastReviewed.IsZero() {
				t.Error("last reviewed should be set")
			}
			if state.NextReview.IsZero() {
				t.Error("next review should be set")
			}
		})
	}
}

func TestIsNew(t *testing.T) {
	t.Run("unreviewed card is new", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		if !store.IsNew("never seen") {
			t.Error("unreviewed card should be new")
		}
	})

	t.Run("reviewed card is not new", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		store.Rate("reviewed card", Good)
		if store.IsNew("reviewed card") {
			t.Error("reviewed card should not be new")
		}
	})
}

func TestIsOverdue(t *testing.T) {
	t.Run("new card is not overdue", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		if store.IsOverdue("new card") {
			t.Error("new card should not be overdue")
		}
	})

	t.Run("card due more than one day ago is overdue", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		key := CardKey("old card")
		store.Cards[key] = CardState{
			NextReview: time.Now().Add(-72 * time.Hour),
			EaseFactor: 2.5,
		}
		if !store.IsOverdue("old card") {
			t.Error("card due 3 days ago should be overdue")
		}
	})

	t.Run("recently due card is not overdue", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		key := CardKey("recent card")
		store.Cards[key] = CardState{
			NextReview: time.Now().Add(-1 * time.Hour),
			EaseFactor: 2.5,
		}
		if store.IsOverdue("recent card") {
			t.Error("card due 1 hour ago should not be overdue")
		}
	})
}

func TestStreak(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		if got := store.Streak(); got != 0 {
			t.Errorf("streak = %d, want 0", got)
		}
	})

	t.Run("single day today", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		store.Cards["a"] = CardState{
			LastReviewed: time.Now(),
		}
		if got := store.Streak(); got != 1 {
			t.Errorf("streak = %d, want 1", got)
		}
	})

	t.Run("consecutive days", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		now := time.Now()
		store.Cards["a"] = CardState{
			LastReviewed: now,
		}
		store.Cards["b"] = CardState{
			LastReviewed: now.AddDate(0, 0, -1),
		}
		store.Cards["c"] = CardState{
			LastReviewed: now.AddDate(0, 0, -2),
		}
		if got := store.Streak(); got != 3 {
			t.Errorf("streak = %d, want 3", got)
		}
	})

	t.Run("gap breaks streak", func(t *testing.T) {
		store := &Store{Cards: make(map[string]CardState)}
		now := time.Now()
		store.Cards["a"] = CardState{
			LastReviewed: now,
		}
		// Skip yesterday, reviewed 2 days ago
		store.Cards["b"] = CardState{
			LastReviewed: now.AddDate(0, 0, -2),
		}
		if got := store.Streak(); got != 1 {
			t.Errorf("streak = %d, want 1", got)
		}
	})
}

func TestReviewedToday(t *testing.T) {
	store := &Store{Cards: make(map[string]CardState)}
	now := time.Now()

	store.Cards[CardKey("q1")] = CardState{
		LastReviewed: now,
	}
	store.Cards[CardKey("q2")] = CardState{
		LastReviewed: now.AddDate(0, 0, -1),
	}

	got := store.ReviewedToday([]string{"q1", "q2", "q3"})
	if got != 1 {
		t.Errorf("reviewed today = %d, want 1", got)
	}
}

func TestDueCount(t *testing.T) {
	store := &Store{Cards: make(map[string]CardState)}

	// q1 is new (due)
	// q2 has future review (not due)
	store.Cards[CardKey("q2")] = CardState{
		NextReview: time.Now().Add(48 * time.Hour),
	}

	got := store.DueCount([]string{"q1", "q2"})
	if got != 1 {
		t.Errorf("due count = %d, want 1", got)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "state.json")

	store := &Store{
		Cards: make(map[string]CardState),
		path:  path,
	}
	store.Cards["test"] = CardState{EaseFactor: 2.5}

	if err := store.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}
}
