package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNginxParser_Detect(t *testing.T) {
	p := &NginxParser{}
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"valid nginx", `127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`, true},
		{"json line", `{"level":"INFO"}`, false},
		{"plain text", "just some log text", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.Detect(tt.line))
		})
	}
}

func TestNginxParser_Parse(t *testing.T) {
	p := &NginxParser{}

	t.Run("200 response is INFO", func(t *testing.T) {
		line := `127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "127.0.0.1", entry.Fields["remote_addr"])
		assert.Equal(t, 200, entry.Fields["status"])
	})

	t.Run("500 response is ERROR", func(t *testing.T) {
		line := `10.0.0.1 - - [15/Jan/2024:10:00:00 +0000] "GET /crash HTTP/1.1" 500 0`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "ERROR", entry.Level)
	})

	t.Run("401 response is WARN", func(t *testing.T) {
		line := `10.0.0.1 - user [15/Jan/2024:10:00:00 +0000] "DELETE /api HTTP/1.1" 401 89`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "WARN", entry.Level)
	})

	t.Run("no match returns error and preserves raw", func(t *testing.T) {
		line := "not an nginx line"
		entry, err := p.Parse(line)
		assert.Error(t, err)
		assert.Equal(t, line, entry.Raw)
	})
}
