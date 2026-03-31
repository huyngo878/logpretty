package filter

import (
	"testing"
	"time"

	"github.com/huyngo878/logpretty/parser"
	"github.com/stretchr/testify/assert"
)

func entry(level, message string) parser.LogEntry {
	return parser.LogEntry{Level: level, Message: message, Raw: message}
}

func TestFilter_Level(t *testing.T) {
	tests := []struct {
		name       string
		filterLvl  string
		entryLevel string
		want       bool
	}{
		{"INFO passes INFO", "INFO", "INFO", true},
		{"INFO blocks DEBUG", "INFO", "DEBUG", false},
		{"INFO passes WARN", "INFO", "WARN", true},
		{"INFO passes ERROR", "INFO", "ERROR", true},
		{"WARN blocks INFO", "WARN", "INFO", false},
		{"no filter passes all", "", "DEBUG", true},
		{"unknown entry level passes", "INFO", "CUSTOM", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{Level: tt.filterLvl}
			assert.Equal(t, tt.want, f.Match(entry(tt.entryLevel, "msg")))
		})
	}
}

func TestFilter_Keyword(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		message string
		want    bool
	}{
		{"match in message", "timeout", "connection timeout", true},
		{"no match", "timeout", "success", false},
		{"case insensitive", "TIMEOUT", "connection timeout", true},
		{"empty keyword passes all", "", "anything", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{Keyword: tt.keyword}
			assert.Equal(t, tt.want, f.Match(entry("INFO", tt.message)))
		})
	}
}

func TestFilter_Since(t *testing.T) {
	now := time.Now()

	t.Run("recent entry passes", func(t *testing.T) {
		f := Filter{SinceTime: now.Add(-1 * time.Hour)}
		e := parser.LogEntry{Timestamp: now.Add(-30 * time.Minute), Message: "x"}
		assert.True(t, f.Match(e))
	})

	t.Run("old entry blocked", func(t *testing.T) {
		f := Filter{SinceTime: now.Add(-1 * time.Hour)}
		e := parser.LogEntry{Timestamp: now.Add(-2 * time.Hour), Message: "x"}
		assert.False(t, f.Match(e))
	})

	t.Run("zero timestamp always passes", func(t *testing.T) {
		f := Filter{SinceTime: now.Add(-1 * time.Hour)}
		e := parser.LogEntry{Message: "x"} // zero timestamp
		assert.True(t, f.Match(e))
	})

	t.Run("no since filter passes all", func(t *testing.T) {
		f := Filter{}
		e := parser.LogEntry{Timestamp: now.Add(-24 * time.Hour), Message: "x"}
		assert.True(t, f.Match(e))
	})
}
