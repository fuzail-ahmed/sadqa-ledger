# Build Progress — Sadqa Ledger

Append-only build log: one entry per completed phase (see `docs/IMPLEMENTATION_PLAN.md`), newest at the bottom. Lives in `docs/` (not `tmp/`) so it survives `make clean` and travels with forks. Check the last entry to see where the build stands before starting work.

**Keep entries terse — three to six lines maximum. Long entries defeat the purpose.** Template:

```markdown
## Phase N — <name> ✅ <date>
Done: <one line — what now works>
Decisions: <any choice a future contributor would otherwise question>
Gotchas: <anything that cost time or will bite someone again>
Next: <the following phase>
```

## Phase −1 — Planning & tooling ✅ 2026-07-19
Done: six planning docs (PRD, TRD, APP_FLOW, UI_UX_BRIEF, SCHEMA, IMPLEMENTATION_PLAN) plus ACCESSIBILITY, DEPLOY, OPERATOR_RESPONSIBILITIES; legal templates; repo scaffolding (CONTRIBUTING, CI workflow, .gitignore); Makefile with air hot reload.
Decisions: `modernc.org/sqlite` (pure Go, CGO_ENABLED=0); Tailwind v4 standalone CLI + Basecoat vendored via `make setup` — no Node anywhere; languages settled as English + Hindi + Arabic (Urdu dropped).
Gotchas: `modernc.org/sqlite` DSN pragmas use `_pragma=name(value)` — `mattn`-style params are silently ignored, leaving WAL/FKs off.
Next: Phase 0 — Scaffolding.

## Phase 0 — Scaffolding ✅ 2026-07-19
Done: `cmd/server` entrypoint, `internal/config`/`internal/server`, chi router serving one Basecoat-styled templ page, `web/static/embed.go` embedding compiled Tailwind CSS + vendored Basecoat. Verified live on Windows: `make dev` and `make build` both serve `/` at `localhost:8080` with 200s on the page and static assets.
Decisions: Basecoat loaded via its own `<link>`/`<script>`, not merged into `input.css` (`docs/TRD.md` §4).
Gotchas: `.air.toml` built `tmp/main` but ran `tmp/main.exe`, so `make dev` silently failed on Windows — fixed by targeting `tmp/main.exe` explicitly.
Next: Phase 1 — Database & Migrations.

## Phase 1 — Database & Migrations ✅ 2026-07-19
Done: `internal/db` (connection + generic migration runner tracked in `schema_migrations`), `migrations/0001_init.sql` creating all six SCHEMA.md tables/indexes, `cmd/migrate` and `cmd/server` both apply migrations through the same code path, failing fast on error.
Decisions: `SetMaxOpenConns(1)` — single writer keeps SQLite lock handling trivial at this scale; DSN has no `file:` prefix (breaks on Windows drive-letter paths).
Gotchas: none beyond the known `_pragma=` DSN gotcha (already documented) — verified WAL/foreign_keys actually took effect at runtime, not just DSN-string-looks-right.
Verified: fresh DB gets all tables/indexes (`sqlite_master` inspected), restart against an existing DB applies zero migrations, `make lint test build` all pass, no `.db*` files staged.
Next: Phase 2 — Auth & First-Run Setup.

## Phase 2 — Auth & First-Run Setup ✅ 2026-07-19
Done: `internal/auth` (bcrypt, SHA-256-hashed sessions, double-submit CSRF, `RequireAuth` middleware), `/setup` `/login` `/logout` `/admins/new` routes + templ pages, `i18n` package (English only so far), `/` now sits behind auth. Session cleanup runs hourly plus lazily on lookup.
Decisions: CSRF is a stateless double-submit cookie (no new dependency); `SecureCookies` derives from `BASE_URL`'s scheme, no new env var; `/admins/new` is a minimal standalone page, not the full Settings screen (Phase 7) — satisfies "admins can add admins" without building Settings early.
Gotchas: none beyond the known DSN gotcha; test DB config needs a non-zero `SessionLifetime` or sessions expire the instant they're created.
Flagged (not decided silently): no rate-limiting spec found anywhere in docs — skipped dedicated throttling, bcrypt's cost is the only delay (added "login rate limiting / lockout" to the Phase 10 pre-pilot checklist); `/admins/new` is temporary and folds into Settings > Admins in Phase 7.
Verified: `make lint test build` green; manual curl walkthrough of setup → protected page → logout → redirect; raw session token confirmed absent from `sessions` table (test + manual DB check); CSRF-forged POST rejected with 403; touch targets set to 44px (`min-h-11`) on all form inputs/buttons.

## Phase 3 — Core Admin UI Shell ✅ 2026-07-19
Done: shared `AdminShell` component (header + language switch + bottom nav, `details`/`summary` "More" menu) wraps every admin route; all `docs/APP_FLOW.md` §0 admin paths scaffolded with placeholder content; language switch persists per-admin via `admins.language_pref`, adds Hindi and Arabic (RTL) alongside English — all three of the project's committed languages now load, ahead of Phase 9's original schedule.
Decisions: language switcher is plain-HTML submit buttons (no JS) posting to `/lang`, persisted on the admin row rather than a cookie, so it follows an admin across devices; bottom nav follows APP_FLOW's "More" grouping (Contributions/Summary/Settings/Export tucked behind, native `<details>` — no JS popover).
Gotchas: none.
Verified: `make lint test build` green; manual curl pass confirms every scaffolded route 200s and titles resolve/translate correctly; `lang=ar` switch confirmed to flip `<html lang dir>` and translate content; Lighthouse accessibility scored 100 on `/login`, `/`, `/members`, `/contributions/new`, `/expenses`, `/settings`; keyboard reachability confirmed structurally (no `tabindex`/`onclick` divs anywhere — every control is a native `a`/`button`/`summary`), not via a live Tab-key recording (no headless keyboard-automation tool available in this environment).
Next: Phase 4 — Members.

## Phase 4 — Members ✅ 2026-07-19
Done: `internal/members` store (list/search, create, update, activate/reactivate) and `/members`, `/members/new`, `/members/:id/edit`, `/members/:id/toggle` routes replace the Phase 3 placeholders; htmx vendored (first screen needing it) for live name search, returning a fragment on `HX-Request`.
Decisions: members are soft-deleted via `is_active` only, per `docs/SCHEMA.md` — added `created_by_admin_id`/`updated_by_admin_id` columns (migration 0003) mirroring admins' audit-trail precedent; Deactivate's confirmation dialog uses native `confirm()` via a `data-confirm` attribute + one delegated listener in `Shell`, not a dialog library, since templ requires `templ.ComponentScript` for inline `onsubmit` expressions.
Gotchas: templ rejects a plain string expression on `onsubmit`/`onclick` attributes (compile error demanding `templ.ComponentScript`) — the `data-confirm` + listener pattern sidesteps it.
Verified: `make lint test build` green; CRUD + soft-delete covered by `internal/members` unit tests and an `internal/server` HTTP-flow test (add validation error, success toast, search-no-results, deactivate/reactivate, audit columns populated); manual curl walkthrough of setup → add → search → edit → deactivate → reactivate; Lighthouse accessibility scored 100 on `/members`, `/members/new`, `/members/:id/edit`; keyboard reachability confirmed structurally (no `tabindex`/`onclick` divs).
Next: Phase 5 — Contributions & Dashboard.
