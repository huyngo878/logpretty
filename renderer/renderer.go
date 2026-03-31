package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/huyngo878/logpretty/parser"
)

// Options configures the renderer.
type Options struct {
	NoColor bool
	JSON    bool
	Output  io.Writer
}

// Renderer writes formatted log entries to an output writer.
type Renderer struct {
	opts  Options
	theme Theme
}

// New creates a Renderer with the given options.
func New(opts Options) *Renderer {
	if opts.NoColor {
		color.NoColor = true
	}
	return &Renderer{opts: opts, theme: DefaultTheme}
}

// Render formats and writes a parsed log entry.
func (r *Renderer) Render(e parser.LogEntry) error {
	if r.opts.JSON {
		return r.renderJSON(e)
	}
	return r.renderColor(e)
}

// RenderRaw writes an unparsed line with dim styling.
func (r *Renderer) RenderRaw(line string) {
	_, _ = fmt.Fprintln(r.opts.Output, r.theme.Dim.Sprint(line))
}

func (r *Renderer) renderColor(e parser.LogEntry) error {
	var sb strings.Builder

	// Timestamp
	if !e.Timestamp.IsZero() {
		sb.WriteString(r.theme.Timestamp.Sprint(e.Timestamp.Format(time.RFC3339)))
		sb.WriteString(" ")
	}

	// Level badge
	if e.Level != "" {
		sb.WriteString(r.levelColor(e.Level).Sprintf("%-5s", e.Level))
		sb.WriteString(" ")
	}

	// Message
	sb.WriteString(r.levelColor(e.Level).Sprint(e.Message))

	// Fields
	if len(e.Fields) > 0 {
		keys := make([]string, 0, len(e.Fields))
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sb.WriteString(" ")
			sb.WriteString(r.theme.FieldKey.Sprint(k))
			sb.WriteString("=")
			sb.WriteString(r.theme.FieldVal.Sprintf("%v", e.Fields[k]))
		}
	}

	_, err := fmt.Fprintln(r.opts.Output, sb.String())
	return err
}

func (r *Renderer) renderJSON(e parser.LogEntry) error {
	out := map[string]interface{}{
		"message": e.Message,
		"level":   e.Level,
	}
	if !e.Timestamp.IsZero() {
		out["time"] = e.Timestamp.Format(time.RFC3339Nano)
	}
	for k, v := range e.Fields {
		out[k] = v
	}

	b, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("renderer: marshal json: %w", err)
	}
	_, err = fmt.Fprintln(r.opts.Output, string(b))
	return err
}

func (r *Renderer) levelColor(level string) *color.Color {
	switch strings.ToUpper(level) {
	case "TRACE":
		return r.theme.Trace
	case "DEBUG":
		return r.theme.Debug
	case "INFO":
		return r.theme.Info
	case "WARN":
		return r.theme.Warn
	case "ERROR":
		return r.theme.Error
	case "FATAL":
		return r.theme.Fatal
	default:
		return r.theme.Default
	}
}
