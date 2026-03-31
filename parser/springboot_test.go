package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpringBootParser_Detect(t *testing.T) {
	p := &SpringBootParser{}
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"valid springboot", "2024-01-15 10:23:45.123  INFO 12345 --- [main] com.example.App : Started", true},
		{"json line", `{"level":"INFO"}`, false},
		{"nginx line", `127.0.0.1 - - [01/Jan/2024:00:00:00 +0000] "GET / HTTP/1.1" 200 0`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.Detect(tt.line))
		})
	}
}

func TestSpringBootParser_Parse(t *testing.T) {
	p := &SpringBootParser{}

	t.Run("full entry", func(t *testing.T) {
		line := "2024-01-15 10:23:45.123  INFO 12345 --- [main] com.example.Application : Started Application in 3.456 seconds"
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "Started Application in 3.456 seconds", entry.Message)
		assert.Equal(t, "main", entry.Fields["thread"])
		assert.Equal(t, "com.example.Application", entry.Fields["logger"])
		assert.False(t, entry.Timestamp.IsZero())
	})

	t.Run("error level", func(t *testing.T) {
		line := "2024-01-15 10:23:48.012  ERROR 12345 --- [main] com.example.db.DataSource : Connection failed"
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "ERROR", entry.Level)
	})

	t.Run("no match returns error", func(t *testing.T) {
		line := "not a spring boot log"
		entry, err := p.Parse(line)
		assert.Error(t, err)
		assert.Equal(t, line, entry.Raw)
	})
}

func TestCloudWatchParser_Parse(t *testing.T) {
	p := &CloudWatchParser{}

	t.Run("detects and parses cloudwatch", func(t *testing.T) {
		line := "2024-01-15T10:23:45.678Z [ERROR] something went wrong"
		assert.True(t, p.Detect(line))
		entry, err := p.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "ERROR", entry.Level)
		assert.Equal(t, "something went wrong", entry.Message)
		assert.False(t, entry.Timestamp.IsZero())
	})

	t.Run("no match returns error", func(t *testing.T) {
		line := "not a cloudwatch line"
		entry, err := p.Parse(line)
		assert.Error(t, err)
		assert.Equal(t, line, entry.Raw)
	})
}

func TestDetector_Parse(t *testing.T) {
	t.Run("detects json", func(t *testing.T) {
		d := NewDetector("")
		entry, err := d.Parse(`{"level":"INFO","message":"hello"}`)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
	})

	t.Run("detects nginx", func(t *testing.T) {
		d := NewDetector("")
		line := `127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET / HTTP/1.0" 200 100`
		entry, err := d.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
	})

	t.Run("detects springboot", func(t *testing.T) {
		d := NewDetector("")
		line := "2024-01-15 10:23:45.123  WARN 1 --- [main] com.example.App : Low memory"
		entry, err := d.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "WARN", entry.Level)
	})

	t.Run("forced format overrides detection", func(t *testing.T) {
		d := NewDetector("nginx")
		line := `127.0.0.1 - - [15/Jan/2024:10:00:00 +0000] "GET / HTTP/1.1" 200 0`
		entry, err := d.Parse(line)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
	})

	t.Run("unknown forced format falls through", func(t *testing.T) {
		d := NewDetector("unknown")
		entry, err := d.Parse("plain text message")
		require.NoError(t, err)
		assert.NotEmpty(t, entry.Raw)
	})

	t.Run("empty line returns error", func(t *testing.T) {
		d := NewDetector("")
		_, err := d.Parse("")
		assert.Error(t, err)
	})
}
