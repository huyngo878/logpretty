package sanitize

import (
	"regexp"

	"github.com/huyngo878/logpretty/parser"
)

var defaultPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)(token|bearer|auth)\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*\S+`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
}

var privateIPPattern = regexp.MustCompile(
	`\b(10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3})\b`)

// Sanitizer applies redaction patterns to log entries.
type Sanitizer struct {
	redactIPs bool
}

// New returns a Sanitizer. Set redactIPs to true to also redact private IPs.
func New(redactIPs bool) *Sanitizer {
	return &Sanitizer{redactIPs: redactIPs}
}

// Apply redacts sensitive values from an entry's message, raw line, and fields.
func (s *Sanitizer) Apply(e parser.LogEntry) parser.LogEntry {
	e.Message = s.redact(e.Message)
	e.Raw = s.redact(e.Raw)

	for k, v := range e.Fields {
		if str, ok := v.(string); ok {
			e.Fields[k] = s.redact(str)
		}
	}

	return e
}

func (s *Sanitizer) redact(input string) string {
	out := input
	for _, re := range defaultPatterns {
		out = re.ReplaceAllStringFunc(out, func(match string) string {
			// Keep the key name, redact the value portion
			idx := regexp.MustCompile(`[:=]\s*\S+`).FindStringIndex(match)
			if idx != nil {
				return match[:idx[0]] + "=[REDACTED]"
			}
			return "[REDACTED]"
		})
	}

	if s.redactIPs {
		out = privateIPPattern.ReplaceAllString(out, "[REDACTED]")
	}

	return out
}
