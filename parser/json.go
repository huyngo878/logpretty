package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JSONParser handles structured JSON log lines.
type JSONParser struct{}

func (p *JSONParser) Detect(line string) bool {
	s := strings.TrimSpace(line)
	return strings.HasPrefix(s, "{")
}

func (p *JSONParser) Parse(line string) (LogEntry, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return LogEntry{Raw: line}, fmt.Errorf("json parser: %w", err)
	}

	entry := LogEntry{
		Raw:    line,
		Fields: make(map[string]interface{}),
	}

	// Extract timestamp
	for _, key := range []string{"time", "timestamp", "ts", "@timestamp"} {
		if v, ok := raw[key]; ok {
			entry.Timestamp = parseTime(v)
			delete(raw, key)
			break
		}
	}

	// Extract level
	for _, key := range []string{"level", "severity", "lvl"} {
		if v, ok := raw[key]; ok {
			if s, ok := v.(string); ok {
				entry.Level = normalizeLevel(s)
				delete(raw, key)
				break
			}
		}
	}

	// Extract message
	for _, key := range []string{"message", "msg", "text"} {
		if v, ok := raw[key]; ok {
			if s, ok := v.(string); ok {
				entry.Message = s
				delete(raw, key)
				break
			}
		}
	}

	// Everything else goes to Fields
	for k, v := range raw {
		entry.Fields[k] = v
	}

	return entry, nil
}

func parseTime(v interface{}) time.Time {
	switch val := v.(type) {
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		} {
			if t, err := time.Parse(layout, val); err == nil {
				return t
			}
		}
	case float64:
		return time.Unix(int64(val), 0)
	}
	return time.Time{}
}

func normalizeLevel(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TRACE":
		return "TRACE"
	case "DEBUG", "DBG":
		return "DEBUG"
	case "INFO", "INFORMATION":
		return "INFO"
	case "WARN", "WARNING":
		return "WARN"
	case "ERROR", "ERR":
		return "ERROR"
	case "FATAL", "CRITICAL":
		return "FATAL"
	default:
		return strings.ToUpper(s)
	}
}
