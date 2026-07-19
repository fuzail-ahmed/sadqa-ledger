# Sadqa Ledger

A self-hosted, mobile-first ledger for masjid/community sadqa collections — contributions, expenses, and a public transparency page. Built for 2–3 trusted volunteer admins replacing a paper notebook, running free on one small VM.

## Stack
- **Backend:** Go + `net/http` + chi (thin router, no framework)
- **Templates:** templ (compiled, type-safe Go templates)
- **Interactivity:** HTMX — server returns HTML fragments, no SPA
- **CSS:** Tailwind v4 standalone CLI + Basecoat (vendored, no Node/npm anywhere)
- **Database:** SQLite (WAL) via `modernc.org/sqlite` (pure Go), `database/sql`, `CGO_ENABLED=0`

## Read before coding
- `docs/PRD.md` — before questioning scope or adding/cutting a feature
- `docs/TRD.md` — before any stack, dependency, or architecture decision
- `docs/SCHEMA.md` — before touching the database, migrations, sessions, or exports
- `docs/APP_FLOW.md` — before building or changing any screen or route
- `docs/UI_UX_BRIEF.md` — before writing markup, styles, or layout
- `docs/IMPLEMENTATION_PLAN.md` — before starting a phase, to know its deliverables and acceptance criteria
- `docs/ACCESSIBILITY.md` — before shipping any UI, for the per-phase a11y bar
- `docs/DEPLOY.md` — before touching Docker, Caddy, Litestream, or CI release steps
- `docs/OPERATOR_RESPONSIBILITIES.md` — before changing settings, privacy, or legal-facing features

## Commands
- `make dev` — hot-reload dev server (air)
- `make lint` — gofmt + go vet
- `make test` — go test ./...
- `make build` — production binary + minified CSS
- `make help` — full command list

## Non-negotiable rules
- Money is an **integer in minor units** (paise/cents) — never a float, anywhere.
- **No JSON API.** HTMX handlers return HTML fragments; no REST layer, no SPA.
- Plain `database/sql` with placeholders — no ORM, no query builder, never string-concatenated SQL.
- No business logic in templ templates — compute in Go, pass simple data in.
- All user-facing strings live in `i18n/*.json` (en/hi/ar) — never hardcoded in templates.
- `modernc.org/sqlite` DSN uses `_pragma=name(value)` syntax — never copy `mattn/go-sqlite3` examples (they're silently ignored).
- Soft deletes only: contributions/expenses get `deleted_at`; members/admins are deactivated, never deleted. Every write records `recorded_by_admin_id`.
- Never commit `.db` files, `.env`, or secrets.
- Run `make lint test` and fix failures before reporting any task complete.
- Build **one phase per session** (per `docs/IMPLEMENTATION_PLAN.md`); stop and report at the phase boundary.
- Update `docs/PROGRESS.md` at the end of every phase.

`docs/PROGRESS.md` is the build log — check it first to see where the build currently stands.
