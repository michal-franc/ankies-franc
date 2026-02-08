package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIsDeckIgnored(t *testing.T) {
	tests := []struct {
		name        string
		ignoreDecks []string
		deck        string
		want        bool
	}{
		{
			name:        "exact match",
			ignoreDecks: []string{"leetcode"},
			deck:        "leetcode",
			want:        true,
		},
		{
			name:        "prefix match with dot separator",
			ignoreDecks: []string{"leetcode"},
			deck:        "leetcode.dp.tasks",
			want:        true,
		},
		{
			name:        "no match",
			ignoreDecks: []string{"leetcode"},
			deck:        "golang",
			want:        false,
		},
		{
			name:        "empty ignore list",
			ignoreDecks: []string{},
			deck:        "anything",
			want:        false,
		},
		{
			name:        "partial name does not match",
			ignoreDecks: []string{"go"},
			deck:        "golang",
			want:        false,
		},
		{
			name:        "multiple patterns first matches",
			ignoreDecks: []string{"leetcode", "python"},
			deck:        "leetcode",
			want:        true,
		},
		{
			name:        "multiple patterns second matches",
			ignoreDecks: []string{"leetcode", "python"},
			deck:        "python.basics",
			want:        true,
		},
		{
			name:        "multiple patterns none match",
			ignoreDecks: []string{"leetcode", "python"},
			deck:        "golang",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{IgnoreDecks: tt.ignoreDecks}
			got := cfg.IsDeckIgnored(tt.deck)
			if got != tt.want {
				t.Errorf("IsDeckIgnored(%q) = %v, want %v", tt.deck, got, tt.want)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "config.json")

	original := Config{
		NotesPath:   "/home/user/notes",
		IgnoreDecks: []string{"leetcode", "python.basics"},
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("saved file is empty")
	}

	// Round-trip: load back and compare
	var loaded Config
	loadData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if err := json.Unmarshal(loadData, &loaded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if loaded.NotesPath != original.NotesPath {
		t.Errorf("NotesPath = %q, want %q", loaded.NotesPath, original.NotesPath)
	}

	if len(loaded.IgnoreDecks) != len(original.IgnoreDecks) {
		t.Fatalf("IgnoreDecks length = %d, want %d", len(loaded.IgnoreDecks), len(original.IgnoreDecks))
	}

	for i, deck := range loaded.IgnoreDecks {
		if deck != original.IgnoreDecks[i] {
			t.Errorf("IgnoreDecks[%d] = %q, want %q", i, deck, original.IgnoreDecks[i])
		}
	}
}

func TestSaveEmptyIgnoreDecks(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	original := Config{
		NotesPath: "/home/user/notes",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	var loaded Config
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if loaded.NotesPath != original.NotesPath {
		t.Errorf("NotesPath = %q, want %q", loaded.NotesPath, original.NotesPath)
	}

	if len(loaded.IgnoreDecks) != 0 {
		t.Errorf("IgnoreDecks = %v, want empty", loaded.IgnoreDecks)
	}
}
