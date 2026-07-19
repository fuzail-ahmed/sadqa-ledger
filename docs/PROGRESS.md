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
