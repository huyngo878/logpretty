package renderer

import "github.com/fatih/color"

// Theme holds the color attributes for each log level and UI element.
type Theme struct {
	Timestamp *color.Color
	Trace     *color.Color
	Debug     *color.Color
	Info      *color.Color
	Warn      *color.Color
	Error     *color.Color
	Fatal     *color.Color
	Default   *color.Color
	Dim       *color.Color
	FieldKey  *color.Color
	FieldVal  *color.Color
}

// DefaultTheme is the default dark-terminal theme.
var DefaultTheme = Theme{
	Timestamp: color.New(color.FgHiBlack),
	Trace:     color.New(color.FgHiBlack),
	Debug:     color.New(color.FgCyan),
	Info:      color.New(color.FgGreen),
	Warn:      color.New(color.FgYellow),
	Error:     color.New(color.FgRed, color.Bold),
	Fatal:     color.New(color.FgHiRed, color.Bold, color.BgBlack),
	Default:   color.New(color.Reset),
	Dim:       color.New(color.FgHiBlack),
	FieldKey:  color.New(color.FgHiBlue),
	FieldVal:  color.New(color.FgWhite),
}
