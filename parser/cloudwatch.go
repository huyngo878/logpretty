package parser

import (
	"fmt"
	"regexp"
	"time"
)

// CloudWatchParser handles AWS CloudWatch log format.
// Example: 2024-01-15T10:23:45.678Z [ERROR] message here
type CloudWatchParser struct{}

var cloudWatchRe = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z)\s+\[(TRACE|DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL)\]\s+(.+)$`)

func (p *CloudWatchParser) Detect(line string) bool {
	return cloudWatchRe.MatchString(line)
}

func (p *CloudWatchParser) Parse(line string) (LogEntry, error) {
	m := cloudWatchRe.FindStringSubmatch(line)
	if m == nil {
		return LogEntry{Raw: line}, fmt.Errorf("cloudwatch parser: no match for line")
	}

	t, _ := time.Parse(time.RFC3339Nano, m[1])

	return LogEntry{
		Timestamp: t,
		Level:     normalizeLevel(m[2]),
		Message:   m[3],
		Fields:    map[string]interface{}{},
		Raw:       line,
	}, nil
}
