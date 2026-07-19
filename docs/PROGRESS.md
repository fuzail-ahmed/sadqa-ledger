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
Flagged (not decided silently): no rate-limiting spec found anywhere in docs — skipped dedicated throttling, bcrypt's cost is the only delay; `admins` table has no `created_by_admin_id` column in `docs/SCHEMA.md`, so admin creation isn't audit-attributed like contributions/expenses are.
Verified: `make lint test build` green; manual curl walkthrough of setup → protected page → logout → redirect; raw session token confirmed absent from `sessions` table (test + manual DB check); CSRF-forged POST rejected with 403; touch targets set to 44px (`min-h-11`) on all form inputs/buttons.
Next: Phase 3 — Core Admin UI Shell.
