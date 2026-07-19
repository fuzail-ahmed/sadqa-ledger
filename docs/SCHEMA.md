# Database Schema — Sadqa Ledger

SQLite in WAL mode, accessed via `database/sql` with no ORM, using the **`modernc.org/sqlite` driver (pure Go, no cgo)** — see `docs/TRD.md` §5 for the full reasoning. This document is the source of truth for tables, relationships, sessions, permissions, and migrations. Screen-level behavior referencing these tables is in `docs/APP_FLOW.md`.

## 1. Conventions

- All monetary amounts are stored as **integers in minor currency units** (e.g., paise for INR: ₹200.00 → `20000`) to avoid floating-point rounding error in sums — a direct requirement given "accuracy is non-negotiable" (`docs/PRD.md` §3).
- All timestamps are stored as `TEXT` in ISO 8601 UTC (`YYYY-MM-DDTHH:MM:SSZ`), SQLite's recommended format for sortable, comparable datetime strings.
- All tables have an `id INTEGER PRIMARY KEY` (SQLite `rowid` alias), plus `created_at`/`updated_at` audit timestamps where the row can change after creation.
- Soft deletes: rows that a product-level "Delete" affects (contributions, expenses) are never physically removed; a `deleted_at` column marks them. This preserves the audit trail even for corrected/removed entries (`docs/PRD.md` §5, `docs/APP_FLOW.md` §10 Assumptions). Members are deactivated (`is_active`), never deleted, for the same reason.
- Foreign keys are enforced (`PRAGMA foreign_keys = ON` set at connection time — SQLite does not enforce FKs by default). **The exact syntax for setting this depends on the driver — see §1a below.**

### 1a. Connection String / DSN and Pragmas (driver-specific)

`modernc.org/sqlite` uses a **different DSN pragma syntax than `mattn/go-sqlite3`** — this is a real gotcha if anyone copy-pastes a connection string example from elsewhere, so it's called out explicitly here rather than assumed.

`modernc.org/sqlite` accepts a repeatable `_pragma=` query parameter; each occurrence is run as its own `PRAGMA ...` statement at connection time (verified against the driver's own `Driver.Open` documentation on pkg.go.dev):

```
file:sadqa-ledger.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)
```

This is the required connection string form for this project. Do **not** use `mattn/go-sqlite3`-style parameters (e.g. `?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000`) — those are a different driver's convention and are silently ignored (not an error) by `modernc.org/sqlite`, which would leave WAL mode and foreign-key enforcement quietly turned off. At minimum, the connection string should set:

- `_pragma=journal_mode(WAL)` — WAL mode, per `docs/TRD.md` §5.
- `_pragma=foreign_keys(1)` — foreign key enforcement, since SQLite does not enable this by default.
- `_pragma=busy_timeout(5000)` (or similar) — so concurrent access by 2–3 admins plus the public read page waits briefly for a lock rather than failing immediately.

## 2. Entity-Relationship Summary

```
admins (1) ----< contributions   (recorded_by_admin_id)
admins (1) ----< expenses        (recorded_by_admin_id)
admins (1) ----< sessions        (admin_id)

members (1) ----< contributions  (member_id)

group_settings  (single row, group-wide config: privacy toggle, currency, public token, quick-amounts)
```

There is exactly one `group_settings` row per instance, reflecting the "one instance = one masjid" no-multi-tenancy decision (`docs/PRD.md` §10, `docs/TRD.md` §7).

## 3. Tables

### 3.1 `admins`

Admin user accounts (2–3 trusted people, `docs/PRD.md` §4).

```sql
CREATE TABLE admins (
    id              INTEGER PRIMARY KEY,
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,              -- bcrypt hash, see docs/TRD.md §6
    display_name    TEXT NOT NULL,
    language_pref   TEXT NOT NULL DEFAULT 'en',  -- 'en' | 'hi' | 'ar'
    is_active       INTEGER NOT NULL DEFAULT 1,  -- 0/1 boolean; deactivated admins can't log in but audit history is preserved
    created_by_admin_id INTEGER REFERENCES admins(id),  -- NULL for the first/setup-created admin; the acting admin's id otherwise (added migrations/0002_admins_created_by.sql)
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
```

### 3.2 `members`

Community members whose contributions are tracked (`docs/PRD.md` §5.1). Members are data records, not login accounts.

```sql
CREATE TABLE members (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    is_active       INTEGER NOT NULL DEFAULT 1,  -- inactive members excluded from paid/unpaid checklist, history retained
    notes           TEXT,                        -- optional free text (e.g., phone number, household), not surfaced publicly
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_members_active_name ON members (is_active, name);
```

### 3.3 `contributions`

One row per logged payment. Multiple rows per member per month are valid (partial payment + top-up), per `docs/PRD.md` §5.2 and `docs/APP_FLOW.md` §3.

```sql
CREATE TABLE contributions (
    id                    INTEGER PRIMARY KEY,
    member_id             INTEGER NOT NULL REFERENCES members(id),
    amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
    contribution_month    TEXT NOT NULL,          -- 'YYYY-MM', the month the payment counts toward (may differ from created_at, e.g. backdated entry)
    paid_on               TEXT NOT NULL,          -- 'YYYY-MM-DD', actual date payment was received
    recorded_by_admin_id  INTEGER NOT NULL REFERENCES admins(id),
    deleted_at            TEXT,                   -- soft delete; NULL = active row
    deleted_by_admin_id   INTEGER REFERENCES admins(id),
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_contributions_member_month ON contributions (member_id, contribution_month);
CREATE INDEX idx_contributions_month ON contributions (contribution_month) WHERE deleted_at IS NULL;
```

### 3.4 `expenses`

```sql
CREATE TABLE expenses (
    id                    INTEGER PRIMARY KEY,
    description           TEXT NOT NULL,
    amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
    expense_date          TEXT NOT NULL,          -- 'YYYY-MM-DD'
    receipt_photo_path    TEXT,                   -- relative path/key to stored file (local disk or object storage), NULL if none
    recorded_by_admin_id  INTEGER NOT NULL REFERENCES admins(id),
    deleted_at            TEXT,
    deleted_by_admin_id   INTEGER REFERENCES admins(id),
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_expenses_date ON expenses (expense_date) WHERE deleted_at IS NULL;
```

### 3.5 `group_settings`

Single-row table (enforced by application logic + a check constraint on a fixed `id`) holding group-wide configuration referenced throughout `docs/APP_FLOW.md` §7.

```sql
CREATE TABLE group_settings (
    id                       INTEGER PRIMARY KEY CHECK (id = 1),  -- singleton row
    group_name               TEXT NOT NULL,
    currency_code            TEXT NOT NULL DEFAULT 'INR',
    currency_symbol          TEXT NOT NULL DEFAULT '₹',
    show_names_publicly      INTEGER NOT NULL DEFAULT 0,   -- 0/1; default OFF per docs/PRD.md §6
    public_token             TEXT NOT NULL UNIQUE,         -- unguessable random token for /p/:token, see docs/TRD.md §9
    quick_amounts_minor      TEXT NOT NULL DEFAULT '[20000,50000,100000,200000]', -- JSON array of minor-unit amounts for quick-amount chips
    default_public_language  TEXT NOT NULL DEFAULT 'en',
    privacy_policy_url       TEXT,                          -- optional; operator's filled-in privacy policy link, see docs/OPERATOR_RESPONSIBILITIES.md and legal/PRIVACY_POLICY_TEMPLATE.md
    updated_at               TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
```

`public_token` should be generated as a cryptographically random value (e.g., UUIDv4 or 32-byte hex/base62 string) at first-run setup, and regenerable from `/settings` (`docs/APP_FLOW.md` §7) — regeneration is simply an `UPDATE` replacing this value, which immediately invalidates the old public URL.

`privacy_policy_url` is left `NULL` until an operator fills it in via Settings; when set, it's shown as a footer link on both `/settings` and the public page (`docs/APP_FLOW.md` §7, §9). It is not required for the app to function — it's a convenience so an operator's members can always find the operator's own filled-in privacy policy (`legal/PRIVACY_POLICY_TEMPLATE.md`) directly from the app.

### 3.6 `sessions`

Server-side session storage backing the session cookie described in `docs/TRD.md` §6.

```sql
CREATE TABLE sessions (
    token_hash      TEXT PRIMARY KEY,     -- SHA-256 hash (hex) of the session token; the raw token is never stored, see §5
    admin_id        INTEGER NOT NULL REFERENCES admins(id),
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    expires_at      TEXT NOT NULL,
    last_seen_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    user_agent      TEXT,
    ip_address      TEXT
);

CREATE INDEX idx_sessions_admin ON sessions (admin_id);
CREATE INDEX idx_sessions_expires ON sessions (expires_at);
```

Expired sessions are lazily deleted on access and/or cleaned up by a periodic sweep (implementation detail, `docs/IMPLEMENTATION_PLAN.md`). Logout deletes the row immediately (server-side invalidation, not just cookie clearing — see `docs/TRD.md` §9).

Column renamed from `id` to `token_hash` to make explicit that this column never holds a usable, bearer-valid token — see the lookup flow in §5.

## 4. Relationships & Referential Rules

- `contributions.member_id → members.id`: required. A member cannot be hard-deleted while contributions reference it — enforced by not exposing member deletion in the UI at all (only deactivation, §3.2), so this FK is never violated by normal app usage.
- `contributions.recorded_by_admin_id`, `expenses.recorded_by_admin_id`, `*.deleted_by_admin_id → admins.id`: required for creation (not nullable on create), forming the audit trail (`docs/PRD.md` §5). Admin accounts are deactivated, not deleted, for the same referential-integrity reason as members.
- `sessions.admin_id → admins.id`: cascade-relevant on admin deactivation — deactivating an admin should also delete their active sessions (application-level, executed as part of the deactivate-admin action, not a DB-level `ON DELETE CASCADE` since admins are soft-deactivated, not deleted).

## 5. Session Handling (detail)

- **Token generation:** on login, the server generates a cryptographically random session token (e.g., 32 bytes from `crypto/rand`, base64/hex-encoded). This raw token is sent to the browser as the cookie value and is **never** written to the database in raw form.
- **Storage:** the server computes `SHA-256(raw_token)` and stores that hex digest as `sessions.token_hash` (primary key). Storing a hash rather than the raw token means a read-only compromise of the database (e.g., a leaked backup, a copied `.db` file, or a compromised export) cannot be used to impersonate an admin — the attacker would need to reverse a SHA-256 hash, not just read a row.
- **Lookup flow (every authenticated request):**
  1. Read the raw token from the `sadqa_session` cookie.
  2. Compute `SHA-256(raw_token)` in the request handler.
  3. Look up `sessions` by `token_hash = <computed hash>`.
  4. If found and `expires_at` is in the future, treat the request as authenticated as `admin_id`, and update `last_seen_at` (sliding expiration).
  5. If not found, or expired, redirect to `/login` (per `docs/APP_FLOW.md` §10) and clear the cookie.
- **Cookie:** `HttpOnly`, `Secure`, `SameSite=Lax`, name e.g. `sadqa_session`, value = the raw token (not the hash — the hash only ever lives server-side).
- **Expiry:** a configurable session lifetime (default suggestion: 30 days, since admins are few, trusted, and re-logging-in on a personal phone is a minor recurring friction not worth tightening further) refreshed on activity (`last_seen_at` updated, sliding expiration).
- **Logout:** deletes the `sessions` row matching the current token's hash and clears the cookie.
- **Concurrent sessions:** an admin may be logged in on multiple devices simultaneously (one row per device, one hash per device) — no single-session-per-admin restriction, since a real admin may use both a phone and a desktop browser.
- **Why hash instead of storing raw:** this closes the gap the original design left open (see the superseded note in the Assumptions section's history) — hashing session tokens before storage is a standard, low-cost hardening step and is adopted here as the v1 approach rather than deferred.

## 6. Permissions Model

Simple two-tier model, matching `docs/PRD.md` §4 (no per-role granularity in v1):

| Actor | Read | Write |
|---|---|---|
| Any authenticated admin (`admins` row, valid session) | Everything | Everything (members, contributions, expenses, settings, admin management) |
| Public (no session, valid `/p/:token`) | Aggregate totals + activity feed, names only if `group_settings.show_names_publicly = 1` | Nothing |
| Public (no session, no/invalid token) | Nothing | Nothing |

Enforcement is at the HTTP-handler layer: admin routes are wrapped in session-checking middleware (chi middleware, `docs/TRD.md` §3); the public route independently looks up `group_settings.public_token` and never checks for a session, so there is no code path where public and admin permission logic can be accidentally conflated.

## 7. Data Ownership Rules

- All data for one masjid/group lives in one SQLite file, on infrastructure that group's admins control (or a host of their choosing) — no shared central database across groups (`docs/PRD.md` §10, `docs/TRD.md` §7).
- Admins can export the full raw database file and CSVs at any time (`docs/PRD.md` §5.7, `docs/APP_FLOW.md` §8) — the group is never dependent on this software's maintainers to access or migrate their own data.
- **Export exclusions (security, not just convenience):** the on-demand `.db` and CSV exports described in `docs/APP_FLOW.md` §8 must **exclude** the `sessions` table entirely and must **exclude** the `admins.password_hash` column. Concretely:
  - The `.db` download is not a raw `cp` of the live file. It is produced from a sanitized snapshot: either (a) a copy taken via SQLite's backup API with the `sessions` table dropped and `admins.password_hash` nulled/redacted before streaming to the admin, or (b) a fresh SQLite file assembled by re-inserting only the non-sensitive tables/columns. Approach (a) or (b) is an implementation choice for `docs/IMPLEMENTATION_PLAN.md`; the requirement is that the downloaded file never contains a working session token hash or a password hash.
  - The CSV export never touches `sessions` or `admins.password_hash` in the first place, since CSVs are built from `contributions`/`expenses`/`members`/`admins.display_name` only (see §CSV export note below) — no separate redaction step is needed there.
  - Rationale: even though `sessions.token_hash` is now a SHA-256 hash rather than a raw token (§5), and `admins.password_hash` is already a bcrypt hash, neither belongs in a file an admin might casually share, email, or upload elsewhere while trying to back up or migrate their group's financial data. Excluding them entirely removes that risk rather than relying on the hash being hard to reverse.
- Receipt photos are stored either on local disk (under a path outside the web root) or in object storage, referenced by `expenses.receipt_photo_path`; the export function (§ above) should include a means to bundle these files, not just the database rows (see Assumptions).
- Backups (Litestream → R2, `docs/TRD.md` §5) are a separate mechanism from the on-demand export above and are **not** subject to the same exclusion: Litestream replicates the live database file (including `sessions` and `admins`) so a full disaster restore recreates exact application state, including active sessions and admin credentials. This is acceptable because Litestream's R2 destination is a private, credential-gated bucket the admins control (`docs/TRD.md` §9), not a file handed out on request the way an export download is.

## 8. Migration Strategy

- **Tool:** plain, numbered `.sql` migration files (e.g., `migrations/0001_init.sql`, `migrations/0002_add_quick_amounts.sql`), applied in order by a small embedded migration runner in Go (e.g., using a minimal library such as `golang-migrate` or a hand-rolled runner that tracks applied migrations in a `schema_migrations` table). No ORM-driven auto-migration — consistent with the "no ORM" decision (`docs/TRD.md` §5): migrations are explicit, reviewable SQL, appropriate for a financial data schema.
- **Tracking table:**
  ```sql
  CREATE TABLE schema_migrations (
      version     INTEGER PRIMARY KEY,
      applied_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
  );
  ```
- **Application at startup:** the Go binary checks `schema_migrations` on boot and applies any pending migration files embedded in the binary (via `embed`), so a self-hoster upgrading to a new release only needs to restart the container — no manual migration step (important for the "near-zero technical skill" deployment goal, `docs/PRD.md`).
- **Backward compatibility:** migrations should be additive where possible (new nullable columns, new tables) to avoid a scenario where an in-progress deploy briefly runs old code against a new schema; destructive changes (column removal/rename) are done in a two-step migrate-then-clean pattern across releases if ever needed.
- **Seed data:** a first-run check (no rows in `group_settings`) triggers a one-time setup flow (create first admin, set group name/currency, generate `public_token`) rather than a SQL seed file — this is an application-level "first run wizard," detailed in `docs/IMPLEMENTATION_PLAN.md`.

## Assumptions

- CSV export (`docs/APP_FLOW.md` §8) is assumed to export `contributions` and `expenses` as two separate files, joined to `members` names and `admins.display_name` (never `password_hash`, per the export exclusions above), excluding soft-deleted rows by default (with an optional "include deleted" checkbox for full-audit export).
- Receipt photo storage location (local disk vs. object storage) is left as an implementation choice in `docs/IMPLEMENTATION_PLAN.md`; the schema only stores a path/key, keeping the schema agnostic to that choice.
- `contribution_month` is stored as a `'YYYY-MM'` string rather than a proper date or separate year/month integer columns, since it's always used for equality/grouping (never date arithmetic) — string comparison/sorting works correctly for this format and keeps queries simple.
