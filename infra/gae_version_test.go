package infra

import (
	"testing"
	"time"
)

func TestWakeUpInterval(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want time.Duration
	}{
		{"unset disables the loop", "", 0},
		{"zero disables the loop", "0", 0},
		{"negative disables the loop", "-1", 0},
		{"non-numeric disables the loop", "abc", 0},
		{"whitespace is trimmed", "  15  ", 15 * time.Minute},
		{"in-range value is honoured", "30", 30 * time.Minute},
		{"max is honoured", "120", 120 * time.Minute},
		{"above max is clamped", "500", 120 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WAKE_UP_MINUTE", tt.env)
			if got := wakeUpInterval(); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
