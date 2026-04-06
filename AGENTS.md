# Agent Instructions

## Package Manager
- Go modules: `go mod download`, `go run .`
- Frontend assets (only when needed): `npm install`, `npx grunt`

## File-Scoped Commands
| Task | Command |
|------|---------|
| Format touched Go files | `gofmt -w path/to/file.go` |
| Test package | `go test ./path/to/pkg` |
| Targeted test | `go test ./path/to/pkg -run TestName` |
| Full Go test sweep | `go test ./...` |

## Commit Attribution
- AI commits MUST include:
```
Co-Authored-By: Codex GPT-5 <noreply@openai.com>
```

## Key Conventions
- Keep handlers explicit: validate method, auth, and redirects in each flow.
- Admin endpoints MUST be guarded by `requiresUserPermission("admin", ...)`.
- Admin read pages SHOULD use `GET`/`HEAD`; state changes MUST be `POST`.
- Shared mutable maps/stores MUST be guarded with `sync.Mutex`/`sync.RWMutex`.
- Use `reportStore.Snapshot()` for template rendering; do not pass live maps.

## Tool Rules
- Search files/text with `rg`/`rg --files` first.
- Read files in slices (for example `sed -n 'start,endp'`) before editing.
- Prefer `apply_patch` for focused single-file edits.
- Run `gofmt` on every modified `.go` file before tests.
- Validate with `go test ./...` before final handoff.
- Do not use destructive git commands (`reset --hard`, checkout file rollback) unless explicitly requested.
