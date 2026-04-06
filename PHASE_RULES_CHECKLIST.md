# Phase Rules Checklist

## Phase 1

- [x] Identify only Phase 1 tasks
- [x] Implement only Phase 1 scope
- [x] Run `go test ./...`
- [x] Fix issues discovered by tests
- [x] Update `README.md`
- [x] Update `PROJECT_RULES.md`
- [x] Update `TASK_LIST.md`
- [x] Confirm stable phase exit criteria
- [x] Commit cleanly
- [x] Push changes

## Notes

- Tests run:
  - `env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go test ./...`
  - `env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go build ./cmd/ghost`
- Phase 1 commit: `c3e007f` (`Phase 1: add CLI fetch foundation`)
- Phase 1 pushed to `origin/main`
