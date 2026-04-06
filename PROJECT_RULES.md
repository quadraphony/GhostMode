# Project Rules

## Architecture Constraints

- `cmd/ghost/main.go` stays thin and only wires the application entrypoint.
- Phase-specific behavior must remain isolated; do not mix unfinished future-phase logic into current packages.
- Business logic belongs under `internal/`.
- Shared data models live in `pkg/types` only when they are useful across package boundaries.
- Prefer the Go standard library; add third-party dependencies only when a phase clearly requires them.

## Coding Rules

- Keep packages small and responsibilities explicit.
- Use typed sentinel errors where they improve CLI messaging and testing.
- Avoid hidden heuristics; document them when introduced.
- Preserve working behavior and avoid unrelated rewrites.
- No placeholder implementations for deferred features.

## CLI Behavior Rules

- The browser is terminal-only.
- The primary entrypoint is `ghost <url>`.
- Errors must be explicit and user-actionable.
- The tool must never claim to support JavaScript execution.
- Normal output should stay readable and free of debug noise.

## Testing Expectations

- Every phase requires tests for its own scope before completion.
- Tests must be deterministic and avoid live network calls.
- Use `httptest` servers for HTTP behavior.
- Add regression tests for every bug fixed.
- Run `go test ./...` before closing a phase.
