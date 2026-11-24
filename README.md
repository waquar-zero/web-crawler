# Concurrent Web Crawler in Go

A command-line web crawler built using Go that concurrently visits webpages, extracts hyperlinks, and recursively discovers pages while respecting configurable crawling limits.

This project was built to explore Go's concurrency primitives, HTTP networking, HTML parsing, synchronization, and context-based cancellation.

> **Project status:** Educational project. Suitable for experimentation and learning, not intended for large-scale or production web crawling.

## Features

- **Concurrent crawling:** Uses goroutines and `sync.WaitGroup` to crawl multiple URLs concurrently.
- **Concurrency limiting:** Uses a buffered channel as a semaphore to limit simultaneous page fetches.
- **Thread-safe URL deduplication:** Uses a mutex-protected map to avoid crawling the same normalized URL more than once.
- **HTML link extraction:** Parses HTML documents and extracts links from anchor elements.
- **Relative URL resolution:** Converts relative links into absolute URLs.
- **URL normalization:** Normalizes URL schemes, hostnames, default ports, and fragments to reduce duplicate requests.
- **Configurable crawl depth:** Limits how deeply the crawler follows links.
- **Maximum page limit:** Limits the number of unique URLs claimed for crawling.
- **Same-domain filtering:** Optionally restricts crawling to the starting hostname.
- **HTTP response validation:** Checks HTTP status codes and content types.
- **Response-size protection:** Rejects HTML responses larger than the configured 2 MiB limit.
- **Request timeouts and cancellation:** Supports overall crawl timeouts and cancellation using `context.Context`.
- **Basic robots.txt support:** Checks crawling permissions using the `robotstxt` package.
- **Request pacing:** Provides configurable per-host request scheduling.
- **Command-line interface:** Supports flags for controlling crawler behavior.

## Technologies Used

- Go
- `net/http` — HTTP requests
- `golang.org/x/net/html` — HTML parsing
- `github.com/temoto/robotstxt` — robots.txt parsing
- `sync` — Mutexes and WaitGroups
- `context` — Cancellation and timeouts
- `flag` — Command-line arguments
- `net/url` — URL parsing and normalization

## Project Structure

```text
web-crawler/
├── main.go
├── crawler.go
├── fetcher.go
├── url_utils.go
├── robots.go
├── rate_limiter.go
├── crawler_test.go
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE
└── README.md
```

### File Responsibilities

| File | Purpose |
|---|---|
| `main.go` | CLI arguments, validation, HTTP client setup, and application entry point |
| `crawler.go` | Recursive concurrent crawling, synchronization, and crawl limits |
| `fetcher.go` | HTTP requests, response validation, and HTML link extraction |
| `url_utils.go` | URL normalization and filtering |
| `robots.go` | robots.txt fetching, parsing, and permission checks |
| `rate_limiter.go` | Per-host request pacing |
| `crawler_test.go` | Go tests |

## Getting Started

### Prerequisites

Install a Go version compatible with the `go.mod` file.

Verify your installation:

```bash
go version
```

### Clone the Repository

```bash
git clone https://github.com/waquar-zero/web-crawler.git
cd web-crawler
```

### Install Dependencies

```bash
go mod download
```

### Run the Crawler

```bash
go run . -url https://example.com -depth 2
```

### Build an Executable

On Linux or macOS:

```bash
go build -o crawler .
./crawler -url https://example.com -depth 2
```

On Windows:

```powershell
go build -o crawler.exe .
.\crawler.exe -url https://example.com -depth 2
```

## Command-Line Options

| Flag | Description |
|---|---|
| `-url` | Starting URL to crawl (required) |
| `-depth` | Maximum crawl depth (default: 3; allowed: 1–5) |
| `-max-pages` | Maximum number of unique URLs claimed for crawling (default: 50) |
| `-same-domain` | Restrict crawling to the starting hostname |
| `-timeout` | Maximum total crawl duration (default: 30s) |
| `-delay` | Minimum scheduled interval between requests to the same host (default: 500ms) |

Run `go run . -h` to see the available options and their current defaults.

### Example Commands

Crawl a website with depth 2:

```bash
go run . -url https://example.com -depth 2
```

Restrict crawling to the starting hostname:

```bash
go run . -url https://example.com -depth 3 -same-domain=true
```

Limit the crawl to 10 unique URLs:

```bash
go run . -url https://example.com -depth 3 -max-pages 10
```

Set a 20-second overall timeout:

```bash
go run . -url https://example.com -depth 3 -timeout 20s
```

Use a one-second request delay:

```bash
go run . -url https://example.com -depth 3 -delay 1s
```

## How It Works

1. The program reads and validates the starting URL and command-line flags.
2. It initializes the HTTP client, crawler state, robots checker, and rate limiter.
3. Each URL is normalized before being checked against the visited map.
4. A mutex protects the visited map and maximum-page check from concurrent access.
5. The crawler checks robots.txt permissions and applies request pacing.
6. A semaphore limits concurrent page fetches.
7. The HTTP fetcher downloads HTML, validates the response, and extracts hyperlinks.
8. Discovered links are filtered and scheduled recursively until the depth or page limit is reached.
9. A WaitGroup ensures that all scheduled crawl tasks finish before the program exits.

Cancellation can interrupt pending work and context-aware HTTP requests.

## Testing

Run the available Go tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

The automated test suite is not yet comprehensive. Additional tests for robots.txt behavior, request pacing, concurrency, and cancellation are planned improvements.

## Known Limitations

This project intentionally prioritizes learning and simplicity over production-scale crawling.

1. **Single-machine architecture:** The crawler runs locally and does not distribute work across multiple machines.
2. **Unbounded goroutine creation:** The semaphore limits concurrent fetches, but the recursive implementation may create many waiting goroutines.
3. **Fixed response-size limit:** HTML responses larger than 2 MiB are skipped, potentially excluding legitimate pages.
4. **No JavaScript rendering:** Links generated dynamically by JavaScript are not discovered.
5. **Basic robots.txt handling:** Robots policy caching, HTTP failure handling, and redirects have limitations and require further hardening.
6. **Best-effort request pacing:** The rate limiter schedules request times but does not strictly guarantee spacing between actual HTTP request starts.
7. **No persistent storage:** Crawled URLs and results are not saved between executions.
8. **No resume support:** Interrupted crawls cannot resume from their previous state.
9. **Limited URL canonicalization:** Some semantically equivalent URLs may still be treated as different pages.
10. **No distributed coordination:** The crawler has no shared task queue, replication, or multi-node coordination.

## Potential Future Improvements

- Bounded worker pool and URL queue
- More comprehensive automated tests
- Persistent storage for crawled pages
- JSON or CSV export
- Resume interrupted crawls
- Improved robots.txt caching and request scheduling
- Crawl statistics and progress reporting
- Distributed crawling across multiple workers

## Responsible Usage

Use the crawler only on websites you are authorized to crawl.

Respect robots.txt rules, website terms, and server capacity. Start with low depth, small page limits, and conservative request delays.

The current robots.txt implementation is basic and should not be treated as a guarantee of full compliance.

