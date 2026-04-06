# Ghost Mode Browser

Ghost Mode Browser is a terminal-first, privacy-oriented web browser written in Go. It fetches web pages over HTTP/HTTPS, strips noise in later phases, and presents readable content without pretending to be a full JavaScript browser.

## Status

The project is in phased development. Phase 1 is focused on CLI foundation, URL normalization, HTTP fetching, redirect handling, and HTML content-type validation.

## Why It Exists

- Fast terminal browsing for content-focused workflows
- Honest handling of non-HTML and JavaScript-heavy sites
- Small, testable architecture with a single-binary target

## Current Features

- `ghost <url>` CLI entrypoint
- URL validation and normalization
- HTTP fetching with timeout and user-agent
- Redirect following with a safety cap
- Clean rejection of unsupported non-HTML content
- Developer-friendly fetch summary output

## Build

This environment uses a direct Go binary instead of the Snap launcher:

```bash
env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go build ./cmd/ghost
```

## Test

```bash
env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go test ./...
```

## Run

```bash
env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go run ./cmd/ghost https://example.com
```

## Limitations

- No HTML parsing yet
- No readable text extraction yet
- No link rendering or navigation yet
- No interactive shell yet
- No JavaScript execution, now or later

## Roadmap

- Phase 1: Foundation and fetching
- Phase 2: HTML parsing and cleaning
- Phase 3: Link extraction and URL resolution
- Phase 4: Terminal rendering
- Phase 5: Interactive navigation
- Phase 6: Readability mode
- Phase 7: Bookmarks and persistent history
- Phase 8: Search integration
- Phase 9: Hardening and polish
