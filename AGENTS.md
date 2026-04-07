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

## Admin Improvement Plan (Planning-Only)

### Goal
- Improve administrator visibility and moderation for pastes and users.
- Introduce a more modern admin UI while preserving existing stack and behavior.

### Scope (V1)
- Add an admin dashboard summary page.
- Improve admin monitoring for pastes and users.
- Keep current admin permission model.
- Keep server-rendered templates (no SPA migration).

### Non-Goals (V1)
- No role model expansion beyond `admin`.
- No major frontend framework rewrite.
- No analytics pipeline redesign.

### Route Plan
- Read routes (`GET`/`HEAD`, admin-guarded):
  - `/admin` (landing + quick links/cards)
  - `/admin/dashboard` (summary metrics)
  - `/admin/reports` (existing, modernized UI)
  - `/admin/pastes` (paste monitoring list + filters)
  - `/admin/users` (user monitoring list)
- Mutation routes (`POST`, admin-guarded):
  - `/admin/promote` (existing)
  - `/admin/paste/{id}/delete` (existing)
  - `/admin/paste/{id}/clear_report` (existing)
  - Optional follow-up: demote/disable actions, only after explicit approval.

### Security and Handler Constraints
- Every admin route MUST use `requiresUserPermission("admin", ...)`.
- Admin read pages SHOULD remain `GET`/`HEAD`; state changes MUST remain `POST`.
- Keep handlers explicit for method checks, auth checks, and redirects.

### Data and Concurrency Plan
- Build snapshot/view-model structs for templates:
  - `AdminDashboardSnapshot`
  - `AdminPasteRow`
  - `AdminUserRow`
- Never pass live mutable maps to templates.
- Continue using `reportStore.Snapshot()` for report rendering.
- Guard shared mutable state with `sync.Mutex`/`sync.RWMutex`.

### UI Modernization Plan (Lightweight)
- Add admin card layout and spacing polish in existing stylesheet.
- Add KPI cards: pastes, reports, users, admins, expiring.
- Add clearer list UI: badges, compact actions, filter/search row.
- Add mobile-friendly stacking and full-width primary actions on small screens.
- Keep visual changes incremental and compatible with current templates.

### Implementation Order
1. Add read-model structs and data assembly helpers.
2. Add new admin routes and handlers (dashboard, pastes, users).
3. Add/update templates for dashboard, reports, pastes, users.
4. Apply lightweight stylesheet updates for modern admin presentation.
5. Add/extend route + handler tests.

### Validation Gates
- `gofmt -w` on each touched `.go` file.
- `go test ./...` must pass before handoff.
- Manual check:
  - Admin pages load with admin account.
  - Non-admin access denied for all admin routes.
  - POST actions redirect and flash correctly.

## Admin Improvement Plan - Progress Handover

### Completed
- Added admin read-model and snapshot assembly in `admin.go`:
  - `AdminDashboardSnapshot`
  - `AdminReportRow`
  - `AdminPasteRow`
  - `AdminUserRow`
- Added admin handlers:
  - `adminHomeHandler`
  - `adminDashboardRedirectHandler`
  - `adminReportsHandler`
  - `adminPastesHandler`
  - `adminUsersHandler`
- Added/updated admin routes (all admin-guarded with `requiresUserPermission("admin", ...)`):
  - `GET/HEAD /admin`
  - `GET/HEAD /admin/dashboard` (compatibility redirect to `/admin`)
  - `GET/HEAD /admin/reports`
  - `GET/HEAD /admin/pastes`
  - `GET/HEAD /admin/users`
- Preserved mutation routes as `POST` only:
  - `/admin/promote`
  - `/admin/paste/{id}/delete`
  - `/admin/paste/{id}/clear_report`
- Updated templates:
  - `templates/admin_home.tmpl`
  - `templates/admin_reports.tmpl`
  - Updated `templates/admin_dashboard.tmpl` (legacy template retained in repo)
  - Added `templates/admin_pastes.tmpl`
  - Added `templates/admin_users.tmpl`
- Applied lightweight admin UI improvements in `public/css/master.less`:
  - KPI cards, nav link row, filter row, mobile-friendly layout
- Merged admin landing/dashboard experience into `admin_home`:
  - `/admin` now renders KPI + quick links + recent reports
  - `/admin/dashboard` redirects to `/admin`
- Updated admin nav links to remove Dashboard duplication across reports/pastes/users pages.
- Updated `templates/admin_pastes.tmpl` to table-based UI (aligned with users page):
  - columns: `Paste`, `Details`, `Reports`, `Action`
  - compact row actions with view/delete controls
- Added `.admin-pastes-table` styles in `public/css/master.less` with responsive behavior.
- Added delete redirect support for paste admin list:
  - `redir=pastes` now redirects to `/admin/pastes`
- Formatting and validation completed:
  - `gofmt -w admin.go main.go`
  - `go test ./...` passed

### Remaining / Follow-up
- Add focused tests for new admin handlers/routes (currently no test coverage for these changes).
- Perform manual browser verification against real admin/non-admin sessions.
- Optional Docker follow-up:
  - `docker compose config` currently requires `.env`.
  - Compose warns that `version:` is obsolete in `docker-compose.yml`.

## Admin Improvement Plan - Playwright Validation (2026-04-07)

### Environment Used
- Ran app via Docker Compose on dev port `9111` from `.env` (`PORT=9111`).
- URL under test: `http://127.0.0.1:9111`.

### Test Method
- Used Playwright MCP for browser-driven verification.
- Verified behavior in two auth states:
  - non-admin/anonymous
  - authenticated admin
- For authenticated admin verification, used isolated local test cycle:
  - backup existing `data/accounts` directory
  - run with empty `data/accounts`
  - create first user via auth token login page (`/auth/token`) with username/password so first account gets admin
  - restore original `data/accounts` after validation

### Verified Passes
- Non-admin access denied on all admin read routes with expected permission error message:
  - `GET /admin`
  - `GET /admin/dashboard`
  - `GET /admin/reports`
  - `GET /admin/pastes`
  - `GET /admin/users`
- Admin access succeeds on:
  - `GET /admin`
  - `GET /admin/dashboard`
  - `GET /admin/reports`
  - `GET /admin/pastes`
  - `GET /admin/users`
- Admin users page supports filters (`q`, `admin=1`) and renders expected rows.
- Admin pastes page renders table-based UI with expected headers:
  - `Paste`, `Details`, `Reports`, `Action`
- Logout flow works:
  - `POST /auth/logout` invalidates admin session.
  - Subsequent `GET /admin` is denied.
  - Re-login restores admin access.

### Notable Findings
- Browser console shows repeated CSS load error on all tested pages:
  - `GET /css/icon_effects.css` returns HTML / wrong MIME type (`text/html`) and is refused by browser.
  - This impacts UI polish and should be fixed in asset routing/build output.
- Docker warning persists:
  - `docker-compose.yml` has obsolete `version:` key.

### Follow-up Actions
- Add integration tests for admin route authorization matrix:
  - anonymous/non-admin denied
  - admin allowed
- Add tests for admin filter parameters on users/pastes.
- Fix static asset path/build for `css/icon_effects.css` and re-run browser checks.

## Admin Simplification Plan (Home + Dashboard Merge)

### Objective
- Simplify admin IA by merging Admin Home and Dashboard into a single canonical page.
- Improve users table readability by giving the User column substantially more width than other columns.

### Scope
- Combine `/admin` and `/admin/dashboard` behavior into one page.
- Keep admin routes protected and method semantics unchanged.
- Adjust users table column proportions in template/CSS only (no role-model changes).

### Route/Handler Plan
1. Keep `/admin` as canonical admin landing route (GET/HEAD).
2. Make `/admin/dashboard` a compatibility route:
   - preferred: redirect to `/admin` (HTTP 302/303), or
   - acceptable: alias to same handler as `/admin`.
3. Consolidate handler logic:
   - remove duplicate assembly between home and dashboard handlers.
   - expose one view model for KPI + quick links + recent reports.

### Template Plan
1. Merge content of `templates/admin_home.tmpl` and `templates/admin_dashboard.tmpl`.
2. Keep one primary template for admin landing (prefer `admin_home`).
3. Update admin page nav links to reflect merged flow.

### Users Table Layout Plan
1. Update `templates/admin_users.tmpl` column emphasis:
   - User column is primary and widest.
   - Role and Persona are compact.
   - Action remains constrained but usable.
2. Update `public/css/master.less` under `.admin-users-table`:
   - Increase `.admin-user-name` width target (~60–70% on desktop).
   - Reduce non-critical column widths.
   - Keep responsive behavior for mobile (`table-layout: auto`, wrap actions safely).

### Constraints
- Do not weaken admin guardrails (`requiresUserPermission("admin", ...)` on all admin routes).
- Keep admin reads on `GET`/`HEAD`; keep all mutations on `POST`.
- Avoid unrelated refactors.

### Validation
- `gofmt -w` on touched `.go` files.
- `go test ./...` passes.
- Manual browser checks:
  - `/admin` renders merged page.
  - `/admin/dashboard` resolves correctly (redirect/alias).
  - Non-admin still denied for all admin pages.
  - Users table visibly prioritizes User column on desktop and remains usable on mobile.
  - Pastes page renders table-based layout consistent with users page.

## Rules To Agents (Execution Contract)

- Do not weaken admin guardrails: every admin endpoint must use `requiresUserPermission("admin", ...)`.
- Keep admin reads on `GET`/`HEAD`; keep state mutation on `POST` only.
- Never pass live mutable maps/stores directly to templates.
- For reports/UI data, always derive template data from snapshots (`reportStore.Snapshot()` or copied view-models).
- Preserve explicit handler flow: method checks, auth checks, and redirects must stay obvious in each path.
- If touching shared mutable state, use `sync.Mutex`/`sync.RWMutex`.
- Before handoff:
  - Run `gofmt -w` on all touched `.go` files.
  - Run `go test ./...`.
- Prefer scoped changes; do not refactor unrelated parts.
- Do not run destructive git operations unless explicitly requested.
