# logpretty

A pipe-friendly CLI tool that parses, filters, and colorizes raw log output from stdin or a file. Supports JSON, plaintext, Nginx, Spring Boot, and AWS CloudWatch log formats — with automatic format detection, level-based filtering, and built-in secrets redaction.

---

## Why logpretty?

Raw log streams are hard to read. `logpretty` takes messy, noisy log output and turns it into clean, color-coded, scannable output in your terminal — without changing your existing logging setup.

```
# Before
{"time":"2024-01-15T10:23:48Z","level":"ERROR","message":"Database connection failed","error":"timeout after 30s","retries":3}

# After  (shown in bold red)
2024-01-15T10:23:48Z ERROR Database connection failed  error=timeout after 30s  retries=3
```

---

## Installation

### Download binary (recommended)

Download the latest release for your platform from the [GitHub Releases](https://github.com/huyngo878/logpretty/releases) page, then add it to your `PATH`.

```bash
# macOS / Linux (amd64)
curl -L https://github.com/huyngo878/logpretty/releases/latest/download/logpretty_linux_amd64.tar.gz | tar xz
sudo mv logpretty /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/huyngo878/logpretty/releases/latest/download/logpretty_darwin_arm64.tar.gz | tar xz
sudo mv logpretty /usr/local/bin/
```

Verify the download using the `checksums.txt` file attached to each release.

### Build from source

Requires Go 1.25+.

```bash
git clone https://github.com/huyngo878/logpretty.git
cd logpretty
make build
# Binary is at dist/logpretty
```

---

## Usage

```
logpretty [file] [flags]
```

Read from **stdin** (pipe mode):

```bash
cat app.log | logpretty
docker logs myapp | logpretty
kubectl logs -f my-pod | logpretty
tail -f /var/log/nginx/access.log | logpretty
```

Read from a **file**:

```bash
logpretty app.log
logpretty /var/log/syslog
```

---

## Flags

| Flag | Description |
|---|---|
| `--level LEVEL` | Show only entries at or above this level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`) |
| `--filter KEYWORD` | Show only lines containing this keyword (case-insensitive) |
| `--since DURATION` | Show only entries from the last duration (e.g. `1h`, `30m`, `15s`) |
| `--format FORMAT` | Force a specific parser: `json`, `plaintext`, `nginx`, `springboot`, `cloudwatch` |
| `--json` | Output as clean JSON (one object per line) — useful for piping to `jq` |
| `--output FILE` | Write output to a file instead of stdout |
| `--no-color` | Disable color output |
| `--redact-ips` | Redact private IP addresses from output |

---

## Examples

### Filter by log level

```bash
# Show only warnings and above
cat app.log | logpretty --level WARN

# Show only errors from a running Docker container
docker logs myapp 2>&1 | logpretty --level ERROR
```

### Search for a keyword

```bash
# Show only lines mentioning "timeout"
cat app.log | logpretty --filter timeout

# Combine level and keyword filters
cat app.log | logpretty --level WARN --filter database
```

### Show recent logs only

```bash
# Last 30 minutes
cat app.log | logpretty --since 30m

# Last hour
cat app.log | logpretty --since 1h
```

### Force a specific format

```bash
# Nginx access logs
tail -f /var/log/nginx/access.log | logpretty --format nginx

# Spring Boot
tail -f spring.log | logpretty --format springboot

# AWS CloudWatch exported logs
logpretty cloudwatch-export.log --format cloudwatch
```

### Pipe clean JSON to jq

```bash
# Extract all error messages
cat app.log | logpretty --json | jq 'select(.level == "ERROR") | .message'

# Write filtered JSON to a file
cat app.log | logpretty --level ERROR --json --output errors.json
```

### Write to a file

```bash
# Save colorless parsed output to a file
cat app.log | logpretty --no-color --output parsed.log
```

---

## Supported Log Formats

`logpretty` auto-detects the format of each line. Use `--format` to override.

### JSON

Any line starting with `{`. Handles common key variants automatically:

| Field | Detected keys |
|---|---|
| Timestamp | `time`, `timestamp`, `ts`, `@timestamp` |
| Level | `level`, `severity`, `lvl` |
| Message | `message`, `msg`, `text` |

```json
{"time":"2024-01-15T10:23:45Z","level":"INFO","message":"Server started","port":8080}
{"ts":1705312800,"lvl":"debug","msg":"Worker started","worker_id":3}
```

Malformed JSON lines are printed as-is in dim gray — `logpretty` never crashes on bad input.

### Nginx (combined access log)

```
127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /index.html HTTP/1.0" 200 612
```

HTTP status codes are mapped to levels: `2xx/3xx` → `INFO`, `4xx` → `WARN`, `5xx` → `ERROR`.

### Spring Boot

```
2024-01-15 10:23:45.123  INFO 12345 --- [main] com.example.App : Started App in 3.456 seconds
```

### AWS CloudWatch

```
2024-01-15T10:23:45.678Z [ERROR] Lambda function timed out after 30 seconds
```

### Plaintext

Catch-all for anything that doesn't match the above. `logpretty` extracts a leading timestamp and the first recognized level keyword (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`).

```
2024-01-15 10:23:45 ERROR connection refused to 10.0.0.5:5432
WARNING: disk usage above 90%
Just a plain message with no structure
```

---

## Secrets Redaction

`logpretty` automatically redacts common secrets from **all output paths** (terminal, `--output` file, `--json` pipe). Redacted values are replaced with `[REDACTED]` — never asterisks, which leak value length.

| Pattern | Example |
|---|---|
| API keys | `api_key=abc123`, `apikey=xyz` |
| Tokens / Auth headers | `token=eyJ...`, `bearer=...`, `auth=secret` |
| Passwords | `password=hunter2`, `pwd=secret` |
| AWS access keys | `AKIAIOSFODNN7EXAMPLE` |
| Private IPs *(opt-in)* | `192.168.1.100`, `10.0.0.5` — enable with `--redact-ips` |

```bash
# Input
{"level":"ERROR","message":"connect failed","password":"hunter2","api_key":"s3cr3t"}

# Output
ERROR connect failed  api_key=[REDACTED]  password=[REDACTED]
```

---

## Color Scheme

| Level | Color |
|---|---|
| `TRACE` | Dark gray |
| `DEBUG` | Cyan |
| `INFO` | Green |
| `WARN` | Yellow |
| `ERROR` | Bold red |
| `FATAL` | Bold red on black |
| Unparsed lines | Dim gray |
| Field keys | Blue |
| Field values | White |

Use `--no-color` to disable all color output.

---

## Pro Features

Some features require a **logpretty pro** license key. Store it in `~/.logpretty/config.yaml`:

```yaml
license_key: your-license-key
profiles:
  production:
    level: WARN
    filter: payment
    since: 1h
  debug-api:
    level: DEBUG
    format: json
    filter: /api/
```

Use a saved profile:

```bash
logpretty --profile production
logpretty app.log --profile debug-api
```

If no valid license key is present, `logpretty` degrades gracefully and prints:

```
This feature requires logpretty pro. See https://logpretty.dev
```

---

## CI / Non-interactive Use

Use `--no-color` to suppress ANSI escape codes in CI environments:

```bash
cat build.log | logpretty --no-color --level ERROR
```

---

## License

MIT
