package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// NginxParser handles Nginx combined access log format.
// Example: 127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 612
type NginxParser struct{}

var nginxRe = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+)`)

func (p *NginxParser) Detect(line string) bool {
	return nginxRe.MatchString(line)
}

func (p *NginxParser) Parse(line string) (LogEntry, error) {
	m := nginxRe.FindStringSubmatch(line)
	if m == nil {
		return LogEntry{Raw: line}, fmt.Errorf("nginx parser: no match for line")
	}

	t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", m[2])
	status, _ := strconv.Atoi(m[6])
	bytes, _ := strconv.Atoi(m[7])

	level := "INFO"
	switch {
	case status >= 500:
		level = "ERROR"
	case status >= 400:
		level = "WARN"
	}

	return LogEntry{
		Timestamp: t,
		Level:     level,
		Message:   fmt.Sprintf("%s %s %s %d", m[3], m[4], m[5], status),
		Fields: map[string]interface{}{
			"remote_addr": m[1],
			"method":      m[3],
			"path":        m[4],
			"protocol":    m[5],
			"status":      status,
			"bytes":       bytes,
		},
		Raw: line,
	}, nil
}
