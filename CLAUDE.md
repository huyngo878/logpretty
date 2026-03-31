# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Overview

**Name:** `logpretty`
**Language:** Go
**Type:** CLI tool — parse, filter, and colorize raw logs (JSON, plaintext, Docker, Spring Boot, Nginx, CloudWatch) from stdin or file input.
**Distribution:** Single static binary via GitHub Releases + Homebrew tap (future)

---

## Local Toolchain (Virtual Environment)

Go and related tools are installed locally — **no system-wide installation required**. The toolchain lives at:

```
%LOCALAPPDATA%\go-sdk\go\          ← Go 1.23.4 runtime (bootstrap; Go 1.25 toolchain is auto-downloaded by `go mod tidy`)
```

Before running any Go commands in a new shell, activate the toolchain:

```bash
export GOROOT="$LOCALAPPDATA/go-sdk/go"   # or C:/Users/<you>/AppData/Local/go-sdk/go
export PATH="$GOROOT/bin:$HOME/go/bin:$PATH"
```

`golangci-lint` and `govulncheck` are installed to `~/go/bin/` via:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

If any tool is missing, re-run the install command above. CI/CD (GitHub Actions) installs its own toolchain and does not depend on the local SDK directory.

---

## Commands

```bash
make build       # Build for current OS
make test        # Run tests with race detector + coverage
make lint        # Run golangci-lint
make vuln        # Run govulncheck
make release-dry # Dry run GoReleaser (no publish)
make clean       # Remove build artifacts
```

Run benchmarks before any PR touching `parser/` or `filter/`:
```bash
go test -bench=. -benchmem ./parser/...
go test -bench=. -benchmem ./filter/...
```

Flag any regression > 10% in the PR description.

---

## Architecture

**Data flow:** Input → Format Detection → Parsing → Sanitization → Filtering → Rendering → Output

**Key rule:** Each stage is strictly isolated:
- All parsing lives in `parser/` — zero parsing in `cmd/`
- All filtering lives in `filter/` — no filter logic in `renderer/`
- `renderer/` receives only already-parsed, already-filtered entries
- `internal/sanitize` runs at parse time — redacted data never reaches the renderer

**Parser interface** (defined in `parser/detector.go`):
```go
type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
    Fields    map[string]interface{}
    Raw       string
}

type Parser interface {
    Detect(line string) bool
    Parse(line string) (LogEntry, error)
}
```

**Error handling:**
- Never crash on malformed input — print as-is with dim style, preserve `Raw`
- All parser errors return `(LogEntry, error)` — callers decide behavior
- Wrap errors with context: `fmt.Errorf("json parser: %w", err)` — never bare `errors.New` inside packages
- Never use `os.Exit` inside packages — only in `main.go`

**Performance:**
- Process line-by-line via `bufio.Scanner` — never load entire file into memory
- For `--watch` / tail mode: goroutines with a done channel, never busy-wait
- Filters run before rendering — don't colorize lines that will be discarded
- Target: 10,000 lines/sec minimum

---

## Security Rules

**Secrets redaction** — `internal/sanitize` must be applied to ALL output paths (terminal, file, JSON pipe):
- API keys: `(?i)(api[_-]?key|apikey)\s*[:=]\s*\S+`
- Tokens: `(?i)(token|bearer|auth)\s*[:=]\s*\S+`
- Passwords: `(?i)(password|passwd|pwd)\s*[:=]\s*\S+`
- AWS keys: `AKIA[0-9A-Z]{16}`
- Redact with `[REDACTED]` — never asterisks (length leaks info)

**Dependencies:**
- No CGO dependencies — binary must be fully static
- Run `govulncheck ./...` before adding any dependency; CI blocks on HIGH/CRITICAL CVEs
- Pin all versions in `go.sum` — no floating versions

**Binary build flags** (enforced in `.goreleaser.yml`):
- `CGO_ENABLED=0`, `-trimpath`, `-ldflags="-s -w"`

---

## Testing Standards

- Every parser must have tests against files in `testdata/`
- Every filter must have table-driven tests with `t.Run()` subtests
- Test malformed input explicitly — the "never crash" rule must be tested
- CI enforces 70% minimum coverage

```go
func TestJSONParser_MalformedLine(t *testing.T) {
    p := &JSONParser{}
    entry, err := p.Parse("{not valid json")
    assert.Error(t, err)
    assert.Equal(t, "{not valid json", entry.Raw)
}
```

---

## Pro Features

Pro features live in `profile/` and are gated by a license key:
- License validated locally (hash check) — no network call on every run
- Stored in `~/.logpretty/config.yaml`
- Failed license check prints: `"This feature requires logpretty pro. See https://logpretty.dev"`
- Never hard-crash on invalid license — degrade gracefully to free tier

---

## What Claude Should Never Do

- Put business logic in `main.go` or `cmd/root.go` — only flag wiring there
- Skip the sanitizer for any output path
- Break the `Parser` interface contract when adding new formats
- Hardcode credentials, tokens, or paths — use flags or config file
- Ignore a returned error — always handle or wrap and return it
- Add a dependency without verifying it passes `govulncheck`
