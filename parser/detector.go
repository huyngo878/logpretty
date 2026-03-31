package parser

import (
	"fmt"
	"time"
)

// LogEntry is the canonical parsed representation of a single log line.
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
	Raw       string
}

// Parser detects and parses a log line into a LogEntry.
type Parser interface {
	Detect(line string) bool
	Parse(line string) (LogEntry, error)
}

// Detector selects the appropriate parser for each line.
type Detector struct {
	forced  string
	parsers []Parser
}

// NewDetector returns a Detector. If format is non-empty it locks to that parser.
func NewDetector(format string) *Detector {
	return &Detector{
		forced: format,
		parsers: []Parser{
			&JSONParser{},
			&NginxParser{},
			&SpringBootParser{},
			&CloudWatchParser{},
			&PlainTextParser{},
		},
	}
}

// Parse selects the right parser and returns a LogEntry.
func (d *Detector) Parse(line string) (LogEntry, error) {
	if line == "" {
		return LogEntry{Raw: line}, fmt.Errorf("detector: empty line")
	}

	if d.forced != "" {
		p := d.parserByName(d.forced)
		if p != nil {
			return p.Parse(line)
		}
	}

	for _, p := range d.parsers {
		if p.Detect(line) {
			return p.Parse(line)
		}
	}

	return (&PlainTextParser{}).Parse(line)
}

func (d *Detector) parserByName(name string) Parser {
	switch name {
	case "json":
		return &JSONParser{}
	case "nginx":
		return &NginxParser{}
	case "springboot":
		return &SpringBootParser{}
	case "cloudwatch":
		return &CloudWatchParser{}
	case "plaintext":
		return &PlainTextParser{}
	}
	return nil
}
