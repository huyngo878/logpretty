package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// PlainTextParser handles unstructured log lines with optional timestamp/level.
type PlainTextParser struct{}

var (
	plainTimestampRe = regexp.MustCompile(
		`^(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)\s*`)
	plainLevelRe = regexp.MustCompile(
		`(?i)\b(TRACE|DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|CRITICAL)\b`)
)

func (p *PlainTextParser) Detect(_ string) bool {
	return true // always falls through as a catch-all
}

func (p *PlainTextParser) Parse(line string) (LogEntry, error) {
	if line == "" {
		return LogEntry{Raw: line}, fmt.Errorf("plaintext parser: empty line")
	}

	entry := LogEntry{
		Raw:    line,
		Fields: make(map[string]interface{}),
	}

	rest := line

	// Try to extract leading timestamp
	if m := plainTimestampRe.FindStringIndex(rest); m != nil {
		ts := rest[m[0]:m[1]]
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		} {
			if t, err := time.Parse(layout, strings.TrimSpace(ts)); err == nil {
				entry.Timestamp = t
				break
			}
		}
		rest = rest[m[1]:]
	}

	// Try to extract level keyword
	if m := plainLevelRe.FindStringIndex(rest); m != nil {
		entry.Level = normalizeLevel(rest[m[0]:m[1]])
	}

	entry.Message = strings.TrimSpace(rest)
	return entry, nil
}
