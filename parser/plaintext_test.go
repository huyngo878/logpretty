package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainTextParser_Parse(t *testing.T) {
	p := &PlainTextParser{}

	t.Run("extracts level keyword", func(t *testing.T) {
		entry, err := p.Parse("2024-01-15 10:23:45 ERROR something went wrong")
		require.NoError(t, err)
		assert.Equal(t, "ERROR", entry.Level)
		assert.NotEmpty(t, entry.Message)
	})

	t.Run("no timestamp or level", func(t *testing.T) {
		entry, err := p.Parse("just a plain message")
		require.NoError(t, err)
		assert.Equal(t, "just a plain message", entry.Message)
		assert.True(t, entry.Timestamp.IsZero())
	})

	t.Run("always detects", func(t *testing.T) {
		assert.True(t, p.Detect("anything"))
		assert.True(t, p.Detect(""))
	})

	t.Run("empty line returns error", func(t *testing.T) {
		_, err := p.Parse("")
		assert.Error(t, err)
	})

	t.Run("warning keyword normalized", func(t *testing.T) {
		entry, err := p.Parse("WARNING: disk space low")
		require.NoError(t, err)
		assert.Equal(t, "WARN", entry.Level)
	})
}
