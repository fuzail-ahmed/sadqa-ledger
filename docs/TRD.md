# Technical Requirements Document — Sadqa Ledger

This document explains what Sadqa Ledger is built with and why. Product requirements are in `docs/PRD.md`; screens and flows are in `docs/APP_FLOW.md`; data model detail is in `docs/SCHEMA.md`. This document does not repeat feature descriptions — it covers stack, architecture, deployment, security, and performance.

## 1. Guiding Principle

Every technical decision optimizes for: **runs on one cheap/free machine, survives its admin losing interest, and is operable by someone who is not a professional sysadmin.** This is why the stack favors a single static binary and boring, well-understood technology over anything that requires an orchestration layer, a managed database, or a Node.js build pipeline.

## 2. Stack Summary

| Layer | Choice |
|---|---|
| Language | Go (latest stable) |
| HTTP routing | `net/http` + `chi` |
| Templates | `templ` (type-safe, compiled Go templates) |
| Interactivity | HTMX |
| CSS | Tailwind CSS (standalone CLI) + Basecoat |
| Database | SQLite (WAL mode), `database/sql`, no ORM |
| Backups | Litestream → Cloudflare R2 |
| Auth | Session cookies + bcrypt, no third-party auth |
| Packaging | Single static binary, single Docker image |
| Hosting | Oracle Cloud "always free" VM, behind Caddy |
| CI/CD | GitHub Actions |
| Installability | PWA (manifest + icons) |
| i18n | JSON translation files (English, Hindi, Urdu, Arabic) |
| Infra cost target | $0/month |

## 3. Backend

**Go, `net/http` + chi.** Go compiles to a single static binary with no runtime dependency, which directly serves the "single static binary" deployment goal and means a self-hoster only needs to run one file (or one Docker container). `chi` is a thin, idiomatic router/middleware layer on top of `net/http` — it adds routing groups, URL params, and middleware chaining without pulling in a full framework. We deliberately avoid heavier frameworks (Gin, Echo, Buffalo): the app's routing needs are simple (a few dozen routes), and a thin layer keeps the codebase readable by a contributor who only knows standard Go.

**No JSON API.** The server renders HTML fragments directly (see HTMX below). There is deliberately no REST/JSON API layer, no API versioning, no separate frontend build. This roughly halves the surface area a contributor needs to understand: one request/response cycle, not two (API + SPA state management).

## 4. Frontend

**templ.** Templates are Go functions, checked at compile time — a typo in a template variable is a build error, not a runtime blank page. This matters more than usual here because financial data displayed wrong (e.g., a botched currency template) is a trust problem, not just a cosmetic bug.

**HTMX.** The app needs interactivity (submit a contribution without a full page reload, live-filter the member search, refresh the dashboard checklist) but not app-level client state, routing, or a build step. HTMX lets the server return HTML fragments that swap into the page, driven by attributes on plain HTML elements. This avoids:
- a JS framework build pipeline (webpack/vite/npm) that a non-JS-focused contributor has to maintain,
- client/server data duplication (the same "what does a contribution look like" logic living in both Go and JS),
- a JSON API purely to feed a SPA.

The tradeoff: HTMX apps have less client-side offline capability and less complex client interactions than a full SPA. That's an acceptable tradeoff for a CRUD ledger app used by 2–3 admins, not a deal-breaker.

**Tailwind CSS (standalone CLI) + Basecoat.** The standalone Tailwind CLI is a single downloadable binary that compiles `.css` from class usage — no Node.js, no `package.json`, no `npm install` step in the build or CI pipeline. This matters because "no Node.js in the build" was a fixed requirement: it removes an entire toolchain (and its supply-chain surface and version-drift problems) from a project that a volunteer contributor needs to be able to build with just Go + one CLI binary. Basecoat provides shadcn-style, accessible component markup/classes usable on plain HTML/templ (buttons, cards, forms, dialogs) without requiring React — giving the app a modern, consistent look without a component framework.

**PWA.** A web app manifest and icon set let the app "Add to Home Screen" on Sohail's phone, so it opens like a native app (own icon, no browser chrome) without maintaining a separate mobile codebase. No offline-first service worker caching of write operations is planned for v1 — see Assumptions.

## 5. Database

**SQLite, WAL mode, `database/sql`, no ORM.** For a single-instance app with 2–3 concurrent writers and a public read-only page, SQLite is not a toy choice — it is the right one: zero network hop between app and database, zero separate service to run, patch, or pay for, and a single `.db` file that is trivially backed up, copied, or handed to another admin. WAL (write-ahead logging) mode allows concurrent readers alongside a single writer, which comfortably covers this app's access pattern (occasional admin writes, frequent-but-light public reads). Raw `database/sql` with hand-written SQL (no ORM/query builder) keeps the query logic visible and auditable — appropriate given financial data is involved and the schema is small (see `docs/SCHEMA.md`) — and avoids an ORM's abstraction cost for a project this size.

**Litestream → Cloudflare R2.** SQLite's one real weakness for a "durability is non-negotiable" requirement is that it's a single file on a single disk: if that disk dies, the data is gone. Litestream continuously streams the SQLite WAL to object storage (Cloudflare R2) in near-real time, giving off-machine, point-in-time-recoverable backups without running a second database server. Cloudflare R2 is chosen specifically because it has a free tier with no egress fees, which matters for hitting the $0 infra-cost target — restoring/reading backups doesn't trigger a surprise bill.

## 6. Auth & Sessions

**Session-cookie auth, bcrypt password hashes, no third-party auth service.** Only 2–3 admins ever log in (see `docs/PRD.md` §4), so there is no case for OAuth, SSO, or a hosted auth provider — those add signup friction, external dependency risk, and recurring cost for zero benefit at this scale. A signed, HTTP-only session cookie plus a `bcrypt`-hashed password column in the `admins` table (see `docs/SCHEMA.md`) is the simplest mechanism that is still secure by current standards. Passwords are set/reset by direct database access or an admin-invite flow (see `docs/APP_FLOW.md`) — there is no self-service signup, since admin accounts are provisioned deliberately, not opened to the public.

**Session tokens are stored hashed, not raw.** On login, the server generates a random token and sends it to the browser as the session cookie's value, but only stores `SHA-256(token)` in the `sessions` table (`sessions.token_hash`, see `docs/SCHEMA.md` §5). Every subsequent request hashes the cookie value again and looks up the session by that hash. This is the same principle as bcrypt-hashing passwords, applied to session tokens: a leak of the database file or a backup (e.g., a stolen `.db` snapshot, or R2 credentials) yields only hashes, not tokens an attacker could replay to impersonate an admin. The cost is one extra `SHA-256` computation per authenticated request, which is negligible next to the app's actual workload (§10, Performance).

Session details (cookie flags, expiry, hashing, lookup flow) are specified in `docs/SCHEMA.md` §5.

## 7. Architecture

Single Go binary containing:
- HTTP server (chi router)
- templ-rendered HTML handlers
- Embedded static assets (compiled CSS, icons, manifest) via Go's `embed`
- SQLite database file on local disk (or a mounted volume in Docker)
- Litestream running as a sidecar process (or separate container) replicating that file continuously

```
[ Browser ] <--HTTPS--> [ Caddy ] <--HTTP--> [ Go binary (chi + templ + HTMX responses) ] <---> [ SQLite file (WAL) ]
                                                                                                        |
                                                                                                   [ Litestream ] --> [ Cloudflare R2 ]
```

Caddy terminates TLS and reverse-proxies to the Go process; it also handles automatic HTTPS certificate issuance/renewal (Let's Encrypt), which is otherwise the most common point of failure for a non-expert self-hoster.

**Multi-tenancy is explicitly not supported** (see `docs/PRD.md` §10). One running instance, one SQLite file, one masjid/group. This is a deliberate simplicity choice: a multi-tenant design would require tenant-scoped queries everywhere, tenant-aware sessions, and a shared-fate database where one group's bug or data issue can affect another's — all complexity with no benefit to the stated goal, since the goal is "communities keep custody of their own data" via forking, not via a shared hosted service.

## 8. Deployment

**Docker image + Oracle Cloud "Always Free" VM + Caddy.** The Go binary, Litestream, and Caddy are packaged into a single Docker image (or docker-compose stack) so a self-hoster runs one command to start the whole stack. Oracle Cloud's Always Free tier provides a small VM (ARM Ampere, permanently free, not a time-limited trial) sufficient for this workload, which is what makes the $0 infra-cost target achievable long-term rather than just for a free-trial period. Full step-by-step instructions for a non-expert are in `docs/DEPLOY.md` (Phase 2 deliverable).

**CI/CD via GitHub Actions.** On push/PR: run `go vet`/`go build`/tests, run the Tailwind CLI build, and (on tagged release) build and publish the Docker image. This keeps releases reproducible without a self-hoster needing to build anything themselves — they can pull a published image.

## 9. Security

- **Transport:** HTTPS everywhere via Caddy's automatic TLS. No plaintext HTTP in production.
- **Passwords:** bcrypt with a standard cost factor; never logged, never stored in plaintext.
- **Sessions:** HTTP-only, `Secure`, `SameSite=Lax` cookies; the raw token lives only in the cookie, the server stores and looks up sessions by `SHA-256(token)` (never the raw value), and logout deletes the server-side row immediately (see `docs/SCHEMA.md` §5).
- **CSRF:** HTMX requests include a CSRF token (double-submit cookie or synchronizer token pattern) on all state-changing routes, since cookie-based sessions are vulnerable to CSRF without one.
- **Public token URL:** the public transparency page is reachable by anyone with the link, by design (see `docs/PRD.md` §5). This is security-by-obscurity, not authentication, and that tradeoff is deliberate: it removes any need for member accounts while still gating the link behind a non-guessable, non-indexed random token (not a short PIN, not a sequential ID). Threat model: acceptable exposure is "someone who was sent the link, or who guesses a 122-bit-equivalent random token" — not acceptable is a search engine indexing it, so the page ships with `noindex` headers/meta tags, and regenerate-token support exists for if a link leaks (see `docs/APP_FLOW.md`, `docs/SCHEMA.md`).
- **Audit trail:** every write records the acting admin's ID and a timestamp (see `docs/SCHEMA.md` §Audit columns). This is a product requirement (`docs/PRD.md` §5) and a security property: it makes tampering or mistakes attributable.
- **Input handling:** all SQL via parameterized queries (`database/sql` placeholders) — no string-concatenated SQL, eliminating SQL injection risk by construction. All template output goes through `templ`, which HTML-escapes by default, eliminating most XSS risk by construction.
- **File uploads (receipt photos):** validated by content-type and size limit, stored outside the web root (or served through a handler that doesn't execute uploaded content), filenames randomized to avoid path traversal or overwrite collisions.
- **Backups:** Litestream replication to R2 is itself a security surface (R2 credentials must be kept secret — see `.env.example` in Phase 2). Backup files are the same sensitive financial data as the live database and must be treated with equivalent care (private bucket, not public).
- **On-demand exports vs. backups:** the admin-facing `.db`/CSV export (`docs/APP_FLOW.md` §8) is deliberately narrower than the Litestream backup. Exports exclude the `sessions` table and `admins.password_hash` entirely (see `docs/SCHEMA.md` §7), because an export is a file an admin might download to a phone, email, or hand to someone while migrating hosts — it should never contain credentials or session material, even hashed ones. Litestream's replicated copy, by contrast, is not a file admins casually move around; it's a private, credential-gated bucket used only for disaster recovery, so it replicates the full database including `sessions` and `admins` to allow an exact state restore.

## 10. Performance

The performance bar for this app is deliberately low-effort to hit: a single admin doing occasional writes, and a public page read occasionally by community members, is a workload SQLite and a single small VM handle without any tuning. Specific choices made anyway, because they're nearly free:

- WAL mode avoids reader/writer lock contention (see §5).
- templ templates compile to Go code — no runtime template parsing cost.
- Tailwind's compiled CSS is a single static file, cacheable and small (only classes actually used are included).
- HTMX responses return small HTML fragments, not full pages, keeping mobile data usage low — directly relevant to the mobile-first, "as fast as a notebook" requirement in `docs/PRD.md`.
- No N+1-prone ORM lazy-loading; hand-written queries make the cost of each screen's data access visible and intentional.

No caching layer, CDN, or read-replica is planned — none is needed at this scale, and adding one would violate the "simplest solution that works" rule.

## 11. i18n

Translation strings live in simple JSON files, one per language (English, Hindi, Urdu, Arabic), keyed by string ID, loaded at startup and selected per-request (by a user/browser preference or a UI toggle — detailed in `docs/APP_FLOW.md`). JSON is chosen over `.po`/gettext tooling because it requires no additional build step or library beyond Go's standard `encoding/json`, keeping with the project's low-toolchain philosophy. Arabic and Urdu are right-to-left languages; layout/CSS implications are covered in `docs/UI_UX_BRIEF.md`.

## Assumptions

- No offline-first service worker for write operations in v1: the PWA manifest makes the app installable and app-like, but contributions/expenses still require a live connection to the server. Full offline queuing is deferred as it adds meaningful complexity (conflict resolution, local storage schema) for a use case (admin in an area with no connectivity) not confirmed as a real problem yet.
- Currency is configured once per deployment (an environment variable / settings-table value, e.g., `INR`/`₹`), not user-switchable at runtime — consistent with `docs/PRD.md`'s no-multi-currency exclusion.
- Litestream restore is a manual, documented CLI procedure (see `docs/DEPLOY.md`, Phase 2) rather than an in-app "restore" button, since restore is a rare, high-stakes operation better done deliberately from a terminal than behind a UI button.
- Single Docker image is assumed to bundle the Go binary and Caddy config; Litestream may run as a second process in the same container or a second container in a compose file — the exact packaging is an implementation detail finalized in `docs/IMPLEMENTATION_PLAN.md`.
