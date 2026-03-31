package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/huyngo878/logpretty/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buf() *bytes.Buffer { return &bytes.Buffer{} }

func TestRenderer_JSON(t *testing.T) {
	tests := []struct {
		name     string
		entry    parser.LogEntry
		contains []string
	}{
		{
			name: "full entry",
			entry: parser.LogEntry{
				Timestamp: time.Date(2024, 1, 15, 10, 23, 45, 0, time.UTC),
				Level:     "INFO",
				Message:   "Server started",
				Fields:    map[string]interface{}{"port": 8080},
			},
			contains: []string{`"level":"INFO"`, `"message":"Server started"`, `"port":8080`, `"time"`},
		},
		{
			name: "no timestamp",
			entry: parser.LogEntry{
				Level:   "ERROR",
				Message: "something failed",
				Fields:  map[string]interface{}{},
			},
			contains: []string{`"level":"ERROR"`, `"message":"something failed"`},
		},
		{
			name: "empty fields",
			entry: parser.LogEntry{
				Level:   "DEBUG",
				Message: "ping",
				Fields:  map[string]interface{}{},
			},
			contains: []string{`"level":"DEBUG"`, `"message":"ping"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := buf()
			r := New(Options{JSON: true, NoColor: true, Output: out})
			err := r.Render(tt.entry)
			require.NoError(t, err)
			for _, want := range tt.contains {
				assert.Contains(t, out.String(), want)
			}
		})
	}
}

func TestRenderer_Color_NoColor(t *testing.T) {
	entry := parser.LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 23, 45, 0, time.UTC),
		Level:     "WARN",
		Message:   "disk space low",
		Fields:    map[string]interface{}{"used_pct": 91},
	}

	out := buf()
	r := New(Options{NoColor: true, Output: out})
	err := r.Render(entry)
	require.NoError(t, err)

	line := out.String()
	assert.Contains(t, line, "WARN")
	assert.Contains(t, line, "disk space low")
	assert.Contains(t, line, "used_pct")
}

func TestRenderer_AllLevels(t *testing.T) {
	levels := []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "UNKNOWN"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			out := buf()
			r := New(Options{NoColor: true, Output: out})
			e := parser.LogEntry{Level: level, Message: "test", Fields: map[string]interface{}{}}
			err := r.Render(e)
			require.NoError(t, err)
			assert.Contains(t, out.String(), "test")
		})
	}
}

func TestRenderer_RenderRaw(t *testing.T) {
	out := buf()
	r := New(Options{NoColor: true, Output: out})
	r.RenderRaw("unparseable garbage line")
	assert.Contains(t, out.String(), "unparseable garbage line")
}

func TestRenderer_NoTimestamp(t *testing.T) {
	out := buf()
	r := New(Options{NoColor: true, Output: out})
	e := parser.LogEntry{Level: "INFO", Message: "no time", Fields: map[string]interface{}{}}
	err := r.Render(e)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "no time")
}

func TestRenderer_FieldsSorted(t *testing.T) {
	out := buf()
	r := New(Options{NoColor: true, Output: out})
	e := parser.LogEntry{
		Level:   "INFO",
		Message: "msg",
		Fields:  map[string]interface{}{"zebra": "z", "apple": "a", "mango": "m"},
	}
	err := r.Render(e)
	require.NoError(t, err)

	line := out.String()
	appleIdx := strings.Index(line, "apple")
	mangoIdx := strings.Index(line, "mango")
	zebraIdx := strings.Index(line, "zebra")
	assert.True(t, appleIdx < mangoIdx && mangoIdx < zebraIdx, "fields should be sorted alphabetically")
}
