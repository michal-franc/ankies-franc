package config

import "testing"

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
