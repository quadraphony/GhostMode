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
- [ ] Commit cleanly
- [ ] Push changes

## Notes

- Tests run:
  - `env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go test ./...`
  - `env HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go /snap/go/current/bin/go build ./cmd/ghost`
- Commit and push are currently blocked because `/home/leeroy/ghostmode` is not a Git repository.
