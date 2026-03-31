package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/huyngo878/logpretty/filter"
	"github.com/huyngo878/logpretty/internal/sanitize"
	"github.com/huyngo878/logpretty/parser"
	"github.com/huyngo878/logpretty/renderer"
	"github.com/spf13/cobra"
)

var (
	flagLevel    string
	flagFilter   string
	flagSince    string
	flagOutput   string
	flagJSON     bool
	flagFormat   string
	flagNoColor  bool
	flagRedactIP bool
)

var rootCmd = &cobra.Command{
	Use:   "logpretty [file]",
	Short: "Parse, filter, and colorize logs from stdin or a file",
	Long: `logpretty parses raw logs (JSON, plaintext, Docker, Spring Boot, Nginx, CloudWatch)
and renders them with color. Pipe from stdin or pass a file path as an argument.`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         run,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&flagLevel, "level", "", "Minimum log level to show (DEBUG, INFO, WARN, ERROR)")
	rootCmd.Flags().StringVar(&flagFilter, "filter", "", "Show only lines matching this keyword")
	rootCmd.Flags().StringVar(&flagSince, "since", "", "Show logs since duration ago (e.g. 1h, 30m)")
	rootCmd.Flags().StringVar(&flagOutput, "output", "", "Write output to file")
	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Output as clean JSON (one object per line)")
	rootCmd.Flags().StringVar(&flagFormat, "format", "", "Force log format (json, plaintext, nginx, springboot, cloudwatch)")
	rootCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
	rootCmd.Flags().BoolVar(&flagRedactIP, "redact-ips", false, "Redact private IP addresses from output")
}

func run(cmd *cobra.Command, args []string) error {
	var input io.Reader = os.Stdin

	if len(args) == 1 {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()
		input = f
	}

	var output io.Writer = os.Stdout
	var fileOut *os.File
	if flagOutput != "" {
		var err error
		fileOut, err = os.Create(flagOutput)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer fileOut.Close()
		output = fileOut
	}

	var sinceTime time.Time
	if flagSince != "" {
		d, err := time.ParseDuration(flagSince)
		if err != nil {
			return fmt.Errorf("invalid --since value %q: %w", flagSince, err)
		}
		sinceTime = time.Now().Add(-d)
	}

	f := filter.Filter{
		Level:     flagLevel,
		Keyword:   flagFilter,
		SinceTime: sinceTime,
	}

	san := sanitize.New(flagRedactIP)
	rend := renderer.New(renderer.Options{
		NoColor: flagNoColor,
		JSON:    flagJSON,
		Output:  output,
	})

	det := parser.NewDetector(flagFormat)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := det.Parse(line)
		if err != nil {
			// malformed — print raw with dim style
			rend.RenderRaw(line)
			continue
		}

		entry = san.Apply(entry)

		if !f.Match(entry) {
			continue
		}

		if err := rend.Render(entry); err != nil {
			return fmt.Errorf("render: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	return nil
}
