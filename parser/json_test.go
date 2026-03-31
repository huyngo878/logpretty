package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser_Detect(t *testing.T) {
	p := &JSONParser{}
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"json object", `{"level":"INFO","message":"hello"}`, true},
		{"plain text", "2024-01-01 INFO hello", false},
		{"nginx log", `127.0.0.1 - - [01/Jan/2024:00:00:00 +0000] "GET / HTTP/1.1" 200 0`, false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.Detect(tt.line))
		})
	}
}

func TestJSONParser_Parse(t *testing.T) {
	p := &JSONParser{}

	t.Run("full entry", func(t *testing.T) {
		line := `{"time":"2024-01-15T10:23:45Z","level":"INFO","message":"Server started","port":8080}`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "Server started", entry.Message)
		assert.Equal(t, float64(8080), entry.Fields["port"])
		assert.False(t, entry.Timestamp.IsZero())
	})

	t.Run("alternate keys ts/lvl/msg", func(t *testing.T) {
		line := `{"ts":1705312800,"lvl":"debug","msg":"Worker started"}`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "DEBUG", entry.Level)
		assert.Equal(t, "Worker started", entry.Message)
		assert.Equal(t, time.Unix(1705312800, 0), entry.Timestamp)
	})

	t.Run("malformed json", func(t *testing.T) {
		line := "{not valid json"
		entry, err := p.Parse(line)
		assert.Error(t, err)
		assert.Equal(t, line, entry.Raw)
	})

	t.Run("empty object", func(t *testing.T) {
		line := `{}`
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, line, entry.Raw)
		assert.Empty(t, entry.Message)
	})

	t.Run("level normalization", func(t *testing.T) {
		cases := []struct{ input, want string }{
			{"warning", "WARN"},
			{"ERROR", "ERROR"},
			{"dbg", "DEBUG"},
			{"information", "INFO"},
			{"critical", "FATAL"},
		}
		for _, c := range cases {
			line := `{"level":"` + c.input + `","message":"x"}`
			entry, err := p.Parse(line)
			require.NoError(t, err)
			assert.Equal(t, c.want, entry.Level, "input: %s", c.input)
		}
	})
}
