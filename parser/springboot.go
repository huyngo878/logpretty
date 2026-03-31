package parser

import (
	"fmt"
	"regexp"
	"time"
)

// SpringBootParser handles Spring Boot default log format.
// Example: 2024-01-15 10:23:45.678  INFO 12345 --- [main] com.example.App : Started App in 3.456 seconds
type SpringBootParser struct{}

var springBootRe = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(TRACE|DEBUG|INFO|WARN|ERROR|FATAL)\s+\d+\s+---\s+\[([^\]]+)\]\s+(\S+)\s+:\s+(.+)$`)

func (p *SpringBootParser) Detect(line string) bool {
	return springBootRe.MatchString(line)
}

func (p *SpringBootParser) Parse(line string) (LogEntry, error) {
	m := springBootRe.FindStringSubmatch(line)
	if m == nil {
		return LogEntry{Raw: line}, fmt.Errorf("springboot parser: no match for line")
	}

	t, _ := time.Parse("2006-01-02 15:04:05.000", m[1])

	return LogEntry{
		Timestamp: t,
		Level:     normalizeLevel(m[2]),
		Message:   m[5],
		Fields: map[string]interface{}{
			"thread": m[3],
			"logger": m[4],
		},
		Raw: line,
	}, nil
}
