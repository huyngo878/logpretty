package sanitize

import (
	"testing"

	"github.com/huyngo878/logpretty/parser"
	"github.com/stretchr/testify/assert"
)

func TestSanitizer_Apply(t *testing.T) {
	s := New(false)

	tests := []struct {
		name     string
		message  string
		contains string
		absent   string
	}{
		{
			name:     "redacts api key",
			message:  "request with api_key=supersecret123",
			contains: "[REDACTED]",
			absent:   "supersecret123",
		},
		{
			name:     "redacts bearer token",
			message:  "Authorization: token=eyJhbGciOiJIUzI1NiJ9",
			contains: "[REDACTED]",
			absent:   "eyJhbGciOiJIUzI1NiJ9",
		},
		{
			name:     "redacts password",
			message:  "connect password=hunter2 to db",
			contains: "[REDACTED]",
			absent:   "hunter2",
		},
		{
			name:     "redacts AWS key",
			message:  "using key AKIAIOSFODNN7EXAMPLE",
			contains: "[REDACTED]",
			absent:   "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:     "safe message unchanged",
			message:  "user logged in successfully",
			contains: "user logged in successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := parser.LogEntry{Message: tt.message, Raw: tt.message}
			result := s.Apply(e)
			assert.Contains(t, result.Message, tt.contains)
			if tt.absent != "" {
				assert.NotContains(t, result.Message, tt.absent)
			}
		})
	}
}

func TestSanitizer_RedactIPs(t *testing.T) {
	t.Run("private IPs redacted when enabled", func(t *testing.T) {
		s := New(true)
		e := parser.LogEntry{Message: "connection from 192.168.1.100", Raw: "connection from 192.168.1.100"}
		result := s.Apply(e)
		assert.NotContains(t, result.Message, "192.168.1.100")
		assert.Contains(t, result.Message, "[REDACTED]")
	})

	t.Run("private IPs kept when disabled", func(t *testing.T) {
		s := New(false)
		e := parser.LogEntry{Message: "connection from 192.168.1.100", Raw: "connection from 192.168.1.100"}
		result := s.Apply(e)
		assert.Contains(t, result.Message, "192.168.1.100")
	})
}
