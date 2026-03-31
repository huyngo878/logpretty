package filter

import (
	"strings"
	"time"

	"github.com/huyngo878/logpretty/parser"
)

// levelOrder maps level names to a numeric rank for comparison.
var levelOrder = map[string]int{
	"TRACE": 0,
	"DEBUG": 1,
	"INFO":  2,
	"WARN":  3,
	"ERROR": 4,
	"FATAL": 5,
}

// Filter holds the active filtering criteria.
type Filter struct {
	Level     string
	Keyword   string
	SinceTime time.Time
}

// Match returns true if the entry passes all active filters.
func (f *Filter) Match(e parser.LogEntry) bool {
	if f.Level != "" {
		minRank, ok := levelOrder[strings.ToUpper(f.Level)]
		if ok {
			entryRank, entryOk := levelOrder[strings.ToUpper(e.Level)]
			if entryOk && entryRank < minRank {
				return false
			}
		}
	}

	if f.Keyword != "" {
		keyword := strings.ToLower(f.Keyword)
		if !strings.Contains(strings.ToLower(e.Message), keyword) &&
			!strings.Contains(strings.ToLower(e.Raw), keyword) {
			return false
		}
	}

	if !f.SinceTime.IsZero() && !e.Timestamp.IsZero() {
		if e.Timestamp.Before(f.SinceTime) {
			return false
		}
	}

	return true
}
