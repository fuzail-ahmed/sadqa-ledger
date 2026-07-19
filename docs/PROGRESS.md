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
Done: `cmd/server` entrypoint, `internal/config` + `internal/server` packages, chi router serving one templ page (`web/templates/pages/home.templ`) styled with Basecoat, `web/static/css/input.css` (Tailwind v4 CSS-first, tokens mirroring `docs/UI_UX_BRIEF.md` §2), `web/static/embed.go` embedding compiled CSS + vendored Basecoat into the binary. Verified end-to-end: `templ generate` + Tailwind build + `go build ./cmd/server` + running binary served `/` at `localhost:8080` with the Basecoat card/button classes present and all three static assets (200 OK, correct byte sizes).
Decisions: kept Basecoat's own CDN bundle un-processed by Tailwind (separate `<link>` tag, per `docs/TRD.md` §4) rather than importing it into `input.css`; `internal/config` split out now (just `PORT`) so env-driven settings have one place to grow into for Phase 1+.
Gotchas: `go.mod`'s `go` directive and the `templ`/`air` versions pinned in the Makefile require Go ≥1.25 — this session's sandbox only had Go 1.24.4 available (no network path to go.dev/proxy.golang.org), so verification ran on temporarily-downgraded toolchain/tool versions (Go 1.24.4, templ v0.3.1001) via GitHub-mirror `replace` directives for blocked `golang.org/x/*` paths; `go.mod`/Makefile were left at their real intended versions (Go 1.25.0, templ v0.3.1020) since a real machine has normal network access — just re-run `make setup && make dev` there to confirm. Also couldn't build/install `air` in the sandbox (pulls in Hugo's dependency tree); verified the equivalent manual pipeline (`templ generate` + `make css` + `go build`) instead, which is what `air`'s `pre_cmd` runs anyway.
Next: Phase 1 — Database & Migrations.
