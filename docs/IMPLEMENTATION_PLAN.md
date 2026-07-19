# Implementation Plan — Sadqa Ledger

Phased build sequence, each phase producing a concrete, demoable deliverable with acceptance criteria. Builds on `docs/PRD.md`, `docs/TRD.md`, `docs/APP_FLOW.md`, `docs/UI_UX_BRIEF.md`, and `docs/SCHEMA.md` — this document does not repeat their content, only sequences the work.

## Phase 0 — Project Setup

**Deliverables:**
- Go module initialized, repo structure established (`cmd/`, `internal/`, `migrations/`, `web/templates`, `web/static`).
- chi router serving a placeholder "It works" route.
- templ toolchain wired into the build (generate step in Makefile/build script).
- Tailwind standalone CLI wired into the build, producing `static/css/output.css` from a starter Tailwind config + Basecoat.
- GitHub Actions CI: build + `go vet` on push/PR.
- Dockerfile producing a runnable single-binary image.

**Acceptance criteria:**
- `docker build && docker run` serves the placeholder page over HTTP locally.
- CI passes on a clean PR.
- A contributor can go from `git clone` to a running local server following only `CONTRIBUTING.md` steps (drafted as a stub in this phase, finalized in Phase 2 of the original brief's document set).

## Phase 1 — Database & Migrations

**Deliverables:**
- SQLite connection setup with WAL mode enabled, `PRAGMA foreign_keys = ON`.
- Migration runner (per `docs/SCHEMA.md` §8) with `migrations/0001_init.sql` creating all tables from `docs/SCHEMA.md` §3.
- `schema_migrations` tracking table and startup migration-apply logic.
- First-run setup detection (empty `group_settings`) stubbed (full wizard in Phase 2).

**Acceptance criteria:**
- Fresh container boot creates a new SQLite file with all tables present, verified by a smoke-test script that inspects `sqlite_master`.
- Restarting the binary against an already-migrated database applies zero migrations and starts cleanly (idempotency check).
- Unit tests cover the migration runner applying a 2-file migration sequence in order.

## Phase 2 — Auth & First-Run Setup

**Deliverables:**
- `admins` table wiring, bcrypt password hashing.
- Session-cookie middleware (chi) per `docs/SCHEMA.md` §5 and `docs/TRD.md` §6.
- Login screen (`/login`) per `docs/APP_FLOW.md` §1.
- First-run setup wizard: create first admin, set group name/currency, generate `public_token` (`docs/SCHEMA.md` §8 seed-data note).
- CSRF protection on all state-changing routes (`docs/TRD.md` §9).
- Logout route clearing server-side session.

**Acceptance criteria:**
- On an empty database, visiting any admin route redirects to the first-run wizard, not `/login`.
- After setup, login/logout works end-to-end with a real bcrypt-hashed password.
- A request to a state-changing route without a valid CSRF token is rejected.
- Session expires per configured lifetime and forces re-login (`docs/APP_FLOW.md` §10).

## Phase 3 — Core Admin UI Shell

**Deliverables:**
- Base layout template (header, bottom nav per `docs/UI_UX_BRIEF.md` §5/§7) shared across admin screens.
- Route scaffolding for all admin paths in `docs/APP_FLOW.md` §0 (empty/placeholder content, real logic in later phases).
- Language switcher wired to the JSON i18n files (`docs/TRD.md` §11) with at least English + one other language loading correctly, including RTL layout verification for Arabic.
- Basecoat components integrated (buttons, cards, forms, toasts, dialogs) per `docs/UI_UX_BRIEF.md` §4.

**Acceptance criteria:**
- Every route in the route map renders without error and shows the shared nav/header.
- Switching language changes visible UI text and, for Arabic, mirrors layout direction.
- Touch targets and type sizes spot-checked against `docs/UI_UX_BRIEF.md` §3/§6 on an actual phone-width viewport.
- **Accessibility (see `docs/ACCESSIBILITY.md`):** the shared shell (nav, header, layout) is fully keyboard-navigable with a visible focus indicator on every focusable element; icon-only nav buttons have `aria-label`s; a first Lighthouse/axe pass run on this shell catches structural issues (missing landmarks, heading order) before they're repeated across every later screen.

## Phase 4 — Members

**Deliverables:**
- Member list, add, edit, deactivate/reactivate (`docs/APP_FLOW.md` §4).
- Search/filter (HTMX live filter).

**Acceptance criteria:**
- All states specified in `docs/APP_FLOW.md` §4 (empty, loading, search-no-results, validation error, success) implemented and manually verified.
- Deactivating a member removes them from any "active members" query used later by the paid/unpaid checklist (tested once Phase 5 lands) while preserving their historical rows.
- **Accessibility:** the member list and add/edit form are fully operable by keyboard; form validation errors are programmatically associated with their field (`aria-describedby` or equivalent), not just visually adjacent — see `docs/ACCESSIBILITY.md`.

## Phase 5 — Contributions & Dashboard

**Deliverables:**
- Add Contribution screen (`docs/APP_FLOW.md` §3), including searchable member field, quick-amount chips, month selector, duplicate-entry soft warning.
- Dashboard (`docs/APP_FLOW.md` §2): stat cards, paid/unpaid checklist, recent activity feed.
- Audit fields (`recorded_by_admin_id`) populated on every write.

**Acceptance criteria:**
- Logging a contribution updates the Dashboard's stat cards and checklist without a manual page refresh (HTMX swap) or, at minimum, on next load.
- Core success metric spot check: an admin can log a contribution in well under the time it takes to write the same entry on paper (informal timed test with a real admin, e.g., Sohail — see `docs/PRD.md` §8).
- All states from `docs/APP_FLOW.md` §3 (empty, loading, validation error, duplicate warning, success-and-reset) implemented.
- Soft-delete/edit of a contribution preserves the audit trail (`docs/SCHEMA.md` §3.3).
- **Accessibility:** the Add Contribution form and Dashboard checklist are fully keyboard-operable; the Dashboard's HTMX-refreshed checklist/stat cards use an ARIA live region so a screen reader user is told when the numbers update after a save (`docs/ACCESSIBILITY.md`).

**Pilot readiness note:** Phases 1–5 alone (schema, auth, members, contributions, dashboard) are enough for an admin to replace the *contribution-recording* half of the paper notebook — logging payments and checking who's paid this month, which is the app's core speed metric (`docs/PRD.md` §8). They are **not** yet enough for a full real-world pilot, because the Dashboard's "Total Expenses" and "Current Balance" figures (`docs/APP_FLOW.md` §2) would be structurally incomplete without Phase 6 — a masjid always has expenses (electricity, repairs), and showing a balance that ignores them would misrepresent the group's actual financial position. Given `docs/PRD.md` §3's non-negotiable accuracy requirement, that's not an acceptable pilot state. **Phases 1–6 (through Expenses) are the minimum for a real-world pilot** — at that point contributions, expenses, and an accurate balance are all present, even though the public transparency page, WhatsApp summary, and on-demand export (Phases 7–8) aren't yet available. If an earlier pilot is wanted, it should be scoped explicitly as "contribution-tracking only, balance not yet meaningful" so Sohail and other admins aren't misled by an incomplete number.

## Phase 6 — Expenses

**Deliverables:**
- Expense list and add form (`docs/APP_FLOW.md` §5), including receipt photo upload with validation (size/type limits, `docs/TRD.md` §9).
- Photo storage wiring (local disk or object storage per the Phase 4/5 implementation choice — `docs/SCHEMA.md` §Assumptions).

**Acceptance criteria:**
- Uploading an oversized or wrong-type file produces the inline error specified in `docs/APP_FLOW.md` §5, not a crash or silent failure.
- Expense totals correctly feed into Dashboard's "All-Time Total Expenses" and "Current Balance" figures.
- **Accessibility:** every rendered receipt photo has descriptive alt text auto-generated from the expense's description and date (e.g., "Receipt photo for: Electricity bill, 12 July 2026"), not left blank or generic (`docs/ACCESSIBILITY.md`).

## Phase 7 — Public Transparency Page & Monthly Summary

**Deliverables:**
- `/p/:token` public page (`docs/APP_FLOW.md` §9), enforcing the privacy setting (`docs/PRD.md` §6) and `noindex` headers (`docs/TRD.md` §9).
- Token regeneration in Settings.
- Monthly Summary screen (`docs/APP_FLOW.md` §6) with clipboard copy and privacy-setting-aware content.
- Settings screen sections (`docs/APP_FLOW.md` §7): privacy toggle, public link, group info, admins, language.
- **A configurable "Privacy Policy" link:** a new field in Settings → Group Info where an operator pastes the URL of their filled-in privacy policy (from `legal/PRIVACY_POLICY_TEMPLATE.md`), stored alongside `group_settings` (a new nullable column, e.g. `privacy_policy_url` — see `docs/SCHEMA.md` for where to add it). If set, this link is surfaced in the footer of both the public transparency page and the admin Settings page. If not set, no broken/placeholder link is shown — the footer simply omits it, so an operator who hasn't done this yet doesn't ship a dead link.

**Acceptance criteria:**
- Toggling the privacy setting immediately changes what the public page and a newly generated summary show, without requiring a server restart.
- Regenerating the public token invalidates the old URL immediately (old token returns the "not available" state from `docs/APP_FLOW.md` §9).
- Summary content is verified correct against a manually-computed month total (accuracy check, given `docs/PRD.md` §3's non-negotiable accuracy requirement).
- Public page never renders any admin-only control or data field regardless of privacy setting — verified by hitting the route with no session at all.
- Setting a privacy policy URL in Settings makes it appear on the public page's footer; leaving it blank shows no link at all (not a placeholder or broken link).
- **Accessibility:** the public transparency page — reachable by anyone, with no login — passes a full Lighthouse/axe/keyboard/screen-reader pass per `docs/ACCESSIBILITY.md`'s testing checklist, since it's the one screen guaranteed to be used by people outside the admin group.

## Phase 8 — Backup & Export

**Deliverables:**
- `/export` screen (`docs/APP_FLOW.md` §8): on-demand `.db` snapshot download, CSV export (contributions + expenses, per `docs/SCHEMA.md` §Assumptions).
- Litestream integration: sidecar/second container replicating the SQLite file to Cloudflare R2 (`docs/TRD.md` §5), configured via `.env` variables (finalized in Phase 2 of the original brief's repo-files phase, `.env.example`).

**Acceptance criteria:**
- Manual `.db` and CSV downloads succeed and open correctly in SQLite Browser / a spreadsheet tool respectively.
- A full restore drill is performed at least once: stop the app, delete the local DB file, restore from the R2-replicated Litestream snapshot, restart, confirm no data loss — directly validating the "zero data loss" success metric (`docs/PRD.md` §8).

## Phase 9 — i18n & PWA Completion

**Deliverables:**
- Full translation coverage for all three languages (English, Hindi, Arabic) across every screen, not just the Phase 3 spot check.
- PWA manifest + icon set (`docs/TRD.md` §4), installable to home screen on both Android and iOS Safari.
- Bilingual label treatment (जमा/खर्च/बाकी etc., `docs/UI_UX_BRIEF.md` §8) applied consistently on Dashboard and Summary.

**Acceptance criteria:**
- No untranslated string ("missing key") fallback visible in any of the three languages on any screen.
- App installs to home screen on a real Android phone (primary device target, `docs/PRD.md`) and opens without browser chrome.
- RTL rendering manually verified on the Dashboard, Add Contribution, and Settings screens for Arabic.
- **Accessibility:** text remains usable (no clipping/overlap) at 200% browser zoom on a phone-width viewport, across at least one LTR and one RTL language (`docs/ACCESSIBILITY.md`).

## Phase 10 — Testing & Hardening

**Deliverables:**
- Unit tests for: money arithmetic (sum/balance calculations), migration runner, session expiry logic, privacy-setting enforcement on the public route.
- Integration/end-to-end smoke tests for the core flow: login → add contribution → dashboard reflects it → generate summary → logout.
- Security pass: confirm parameterized queries throughout (no string-concatenated SQL), confirm CSRF coverage on all state-changing routes, confirm file-upload validation, confirm public route never leans on client-side hiding of admin data (`docs/TRD.md` §9).
- **Full accessibility pass across every screen**, running the complete testing checklist from `docs/ACCESSIBILITY.md`: Lighthouse and axe automated scans, a keyboard-only pass through the full core flow, a screen reader spot check (NVDA/VoiceOver/TalkBack), a measured contrast check against the implemented (not just documented) colors, a 200%-zoom check, and a `prefers-reduced-motion` check — not just the informal legibility spot check this phase previously described.

**Acceptance criteria:**
- CI runs the full test suite on every PR and blocks merge on failure.
- Security pass findings are either fixed or explicitly logged as known/accepted (with reasoning) before Phase 11.
- At least one non-technical person completes the full "log a contribution" flow on their own phone without assistance, timing it against the notebook baseline.
- Accessibility pass findings are either fixed or explicitly logged as known/accepted (with reasoning) before Phase 11 — same bar as the security pass above, per `docs/ACCESSIBILITY.md`.

## Phase 11 — Deployment

**Deliverables:**
- Oracle Cloud Always-Free VM provisioned and documented (`docs/DEPLOY.md`, delivered in the repo-files phase).
- Caddy reverse proxy with automatic HTTPS configured in front of the app.
- Docker Compose stack (app + Litestream, and Caddy either in the same stack or as host-level Caddy) deployed and running continuously.
- GitHub Actions release workflow: on tag, build and publish the Docker image.

**Acceptance criteria:**
- The app is reachable over HTTPS at a real domain/subdomain with a valid, auto-renewing certificate.
- A tagged release triggers CI to publish a pullable image, and redeploying via that image (rather than a local build) is confirmed to work — validating that a self-hoster with no Go toolchain can still deploy from `docs/DEPLOY.md` alone.
- Monthly infra cost confirmed at $0 (Oracle Always-Free tier + Cloudflare R2 free tier within usage limits) — validating the cost success metric (`docs/PRD.md` §8).

## Phase 12 — Polish & Launch Readiness

**Deliverables:**
- Visual polish pass against `docs/UI_UX_BRIEF.md` (spacing consistency, color token cleanup, empty-state copy review).
- Repo files from the brief's Phase 2 (README, LICENSE, CONTRIBUTING, SECURITY, issue/PR templates, CHANGELOG) finalized alongside the working app.
- Screenshot capture for README placeholders, using the real running app (not mockups) once visual polish is done.
- v0.1.0 tag cut.

**Acceptance criteria:**
- A first-time visitor to the repo can go from `git clone` (or `docker run`) to a working, logged-in instance using only the README/DEPLOY docs, with no undocumented steps.
- CHANGELOG's Unreleased section is moved to a dated v0.1.0 entry at tag time.

## Cross-Phase Notes

- **Order rationale:** auth and schema come before any UI (a coding agent or contributor should never build screens against an undefined data model or unauthenticated routes); Members before Contributions (contributions reference members); Dashboard is built alongside Contributions rather than earlier, since it has nothing meaningful to show until contributions exist; Public page and Summary come after both Contributions and Expenses exist, since they aggregate both; Backup/Export and Deployment are placed late because they operate on a schema/feature set that should be stable by then, but the Litestream *mechanism* itself could be wired as early as Phase 1 if the team prefers continuous protection from day one (see Assumptions).
- **Definition of done, project-wide:** a phase is not complete until its acceptance criteria are verified, not merely "code written" — this mirrors the accuracy/trust requirements in `docs/PRD.md` §3.

## Assumptions

- Each phase is assumed to map roughly to one PR or a small PR series, not a strict calendar timebox — actual duration depends on contributor availability, which this plan doesn't presume to know.
- Litestream wiring is sequenced at Phase 8 for documentation clarity (grouped with the rest of "durability" features), but nothing prevents standing up Litestream replication as early as Phase 1 in parallel if a contributor wants continuous backup protection during development too — doing so earlier is strictly safer, never harmful to the plan.
- Testing (Phase 10) is placed after most features rather than strictly test-first (TDD), matching a small-team/solo-contributor open-source pace; teams preferring TDD can pull relevant unit tests forward into each feature's phase without changing the plan's overall structure.
