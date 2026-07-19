# App Flow — Sadqa Ledger

This document specifies every screen, navigation path, button behavior, and state (empty/loading/error/success) so an implementer never has to guess. Product rationale is in `docs/PRD.md`; visual design is in `docs/UI_UX_BRIEF.md`; data shapes are in `docs/SCHEMA.md`.

Notation: **[Admin]** = requires login. **[Public]** = no login required.

## 0. Route Map

| Path | Access | Screen |
|---|---|---|
| `/login` | Public (form), redirects if already logged in | Login |
| `/` | Admin | Dashboard |
| `/members` | Admin | Member list |
| `/members/new` | Admin | Add member (modal/page) |
| `/members/:id/edit` | Admin | Edit member |
| `/contributions/new` | Admin | Add contribution |
| `/contributions` | Admin | Contribution history (filterable) |
| `/expenses` | Admin | Expense list |
| `/expenses/new` | Admin | Add expense |
| `/summary` | Admin | Monthly summary generator |
| `/settings` | Admin | Group settings (privacy toggle, currency, admins, token regeneration) |
| `/export` | Admin | Backup & export |
| `/p/:token` | Public | Public transparency page |
| `/logout` | Admin | Logs out, redirects to `/login` |

Bottom navigation (mobile) / side nav (desktop, if ever widened) covers: Dashboard, Add Contribution, Members, Expenses, Summary, Settings — mirroring the priority order in `docs/UI_UX_BRIEF.md`.

## 1. Login — `/login` [Public form]

**Purpose:** authenticate one of the 2–3 admins.

**Elements:** username/email field, password field, "Log in" button, error area above the form.

**States:**
- *Empty (first load):* form only, no error text shown.
- *Loading:* button shows a spinner and disables on submit to prevent double-submit.
- *Error:* wrong credentials → "Incorrect username or password." shown above the form; fields retain the entered username, password field cleared. No indication of which field was wrong, to avoid username enumeration.
- *Success:* redirect to `/` (Dashboard). Session cookie set.
- Already-logged-in users hitting `/login` are redirected straight to `/`.

Password reset in v1 is out of self-service scope: a locked-out admin is reset by another admin via `/settings` (Admins section) or direct DB access. This is acceptable given only 2–3 trusted admins exist.

## 2. Dashboard — `/` [Admin]

**Purpose:** the home screen; answers "what's our status right now" and "who hasn't paid this month" at a glance. This is the screen an admin sees immediately after login and the one they return to via the nav bar's home icon.

**Layout (top to bottom, mobile-first):**
1. Header: group name, current month/year, logout icon.
2. Four stat cards: This Month's Collection, All-Time Total Collected, All-Time Total Expenses, Current Balance. (Bilingual labels — जमा / खर्च / बाकी — per `docs/UI_UX_BRIEF.md`.)
3. Prominent **"+ Add Contribution"** button — the single most-used action in the app, placed above the fold, thumb-reachable.
4. Paid/Unpaid checklist for the current month: every active member listed, with a paid (✓, amount, date) or unpaid (—) indicator. Tapping an unpaid member jumps straight to Add Contribution pre-filled with that member.
5. Recent activity feed: last ~10 contributions/expenses across both types, newest first, each showing who recorded it (audit trail surfaced in the UI, not just the DB).

**States:**
- *Empty (brand-new instance, no members yet):* stat cards show zero values; checklist area replaced with "No members yet — add your first member to get started" + link to `/members/new`.
- *Empty (members exist, none paid yet this month):* checklist shows all members as unpaid; no special empty state needed beyond that.
- *Loading:* skeleton placeholders for stat cards and checklist rows while data loads (relevant mainly on first paint / slow connections, per mobile-first requirement).
- *Error:* if the dashboard query fails (e.g., DB unreachable), show a full-width banner: "Couldn't load the dashboard. [Retry]" — do not show partial/stale numbers as if current.
- *Success:* normal populated view as above.

## 3. Add Contribution — `/contributions/new` [Admin]

**Purpose:** the core success-metric screen — must be faster than writing in a notebook.

**Flow:**
1. Member field: a searchable/type-ahead list of active members (HTMX-powered live filter as the admin types — no full page reload). Recently-paid members are visually de-emphasized but not hidden, since correcting/adding a second payment is valid.
2. Amount field: numeric keypad on mobile (`inputmode="numeric"`). Quick-amount chips above or below the field (e.g., ₹200 / ₹500 / ₹1000 / ₹2000, configurable per `docs/SCHEMA.md` settings) — tapping a chip fills the amount instantly; the admin can still type a custom amount.
3. Month field: defaults to the current month; a simple prev/next or dropdown to log a late/backdated payment for a prior month.
4. **"Save"** button, full-width, thumb-reachable at the bottom (mobile-first placement).

**States:**
- *Empty:* member field unfocused/empty, amount empty, month defaulted to current — form is otherwise ready to use immediately (no unnecessary steps before the first tap).
- *Loading:* on Save, button disables and shows a spinner; the rest of the form is not blocked from being visually inspected.
- *Validation error:* no member selected → inline error under member field, "Select a member." Amount is zero/blank/non-numeric → inline error under amount field, "Enter an amount greater than zero." Errors appear without a full page reload (HTMX fragment swap).
- *Success:* on save, an HTMX-driven success toast/banner ("Saved — ₹500 from Abdul Rahman for July 2026") appears, the form resets (member field cleared, amount cleared, month stays on current) ready for the *next* entry, since admins typically log several contributions in a row. This reset-and-ready behavior is central to the "faster than a notebook" goal.
- *Duplicate-looking entry:* if a contribution already exists for this member+month, do not block the save (see `docs/SCHEMA.md` — multiple rows per member/month are valid, e.g. partial payment + top-up) but show a soft inline notice: "Abdul Rahman already has a payment logged for July 2026 — this will be added as a second entry." Admin can proceed or cancel.

## 4. Members — `/members` [Admin]

**Purpose:** manage the member roster.

**Layout:** search box at top; list of members (name, status badge Active/Inactive, this-month paid indicator); each row has Edit and Deactivate/Reactivate actions; a floating/persistent "+ Add Member" button.

**States:**
- *Empty:* "No members yet." + prominent Add Member button/illustration.
- *Loading:* skeleton rows.
- *Search with no results:* "No members match '<query>'."
- *Success:* populated list, paginated or infinite-scroll if the roster grows large (unlikely at this scale, but not hard-capped).

### 4a. Add/Edit Member — `/members/new`, `/members/:id/edit` [Admin]

**Elements:** name (required), status toggle (Active/Inactive, defaults Active on create), Save/Cancel.

**States:**
- *Validation error:* empty name → inline "Name is required."
- *Success:* saves, returns to `/members` with a confirmation toast ("Member added" / "Member updated"), list reflects the change immediately.
- Deactivating a member does not delete their historical contribution records (see `docs/SCHEMA.md`) — a confirmation dialog on Deactivate explains this: "Farhan will be hidden from the monthly paid/unpaid checklist but their payment history is kept."

## 5. Expenses — `/expenses`, `/expenses/new` [Admin]

**Layout (list):** reverse-chronological list of expenses (date, description, amount, receipt thumbnail if present, recorded-by); "+ Add Expense" button.

**Add Expense form:** description (required), amount (required, > 0), date (defaults to today), optional receipt photo upload (camera or gallery on mobile).

**States:**
- *Empty (list):* "No expenses recorded yet."
- *Loading (photo upload):* upload progress indicator; Save disabled until upload completes or is skipped.
- *Validation error:* missing description/amount → inline errors, same pattern as Add Contribution.
- *Upload error:* photo too large or wrong file type → inline error, "Photo must be under 5MB (JPG/PNG)." — save can proceed without a photo if the admin removes it.
- *Success:* toast confirmation, list updates, form resets for next entry.

## 6. Monthly Summary — `/summary` [Admin]

**Purpose:** generate the WhatsApp-shareable text block described in `docs/PRD.md` §5.

**Flow:** month selector (defaults to current month) → generated summary preview (plain text block, formatted for readability in WhatsApp — line breaks, simple bold via WhatsApp's `*asterisk*` syntax) → "Copy to clipboard" button → confirmation ("Copied!").

**Content of the summary** (respects the privacy setting from `docs/PRD.md` §6):
- Always: month/year, total collected, total spent, closing balance.
- If names-public setting is ON: list of contributing members and amounts.
- If OFF: no names, contributor list omitted entirely (not shown as "Anonymous x12" — simply not included, to avoid implying a headcount that could out low-privacy givers via elimination).

**States:**
- *Empty (no data for selected month):* "No contributions or expenses recorded for [Month]." — summary still generatable (shows zeros) since a "quiet month" is still worth confirming to the group.
- *Loading:* brief spinner while the summary text is composed (should be near-instant given SQLite's speed at this scale).
- *Success (copy):* clipboard copy confirmation toast; a manual "select all" fallback text area is shown if clipboard API access is denied by the browser.

## 7. Settings — `/settings` [Admin]

**Sections:**
1. **Privacy:** "Show contributor names publicly" toggle (default OFF) — see `docs/PRD.md` §6. Changing this takes effect immediately on the public page and next-generated summaries; a confirmation dialog explains the effect before saving ("Turning this on will show member names next to amounts on the public page and in summaries.").
2. **Public link:** displays the current public URL (`/p/:token`) with a Copy button and a "Regenerate link" button (invalidates the old token immediately — used if a link leaks). Regeneration requires a confirm dialog: "The old link will stop working immediately."
3. **Group info:** group/masjid display name, currency symbol/code (see `docs/TRD.md` §Assumptions), quick-amount chip values used on Add Contribution, and an optional **Privacy Policy URL** field where the operator pastes the link to their filled-in privacy policy (from `legal/PRIVACY_POLICY_TEMPLATE.md`). When set, this link appears in the footer of this Settings page and the public transparency page (§9); when blank, no link is shown at all.
4. **Admins:** list of admin accounts (username, last login), add-admin (sets a temporary password an existing admin shares out-of-band), remove-admin, reset-password actions.
5. **Language:** default language for the public page; each admin's own language preference is a separate per-session/per-account choice (see `docs/UI_UX_BRIEF.md` for language switcher placement).

**States:**
- *Success:* each section saves independently (HTMX fragment per section) with its own inline confirmation — avoids a single giant form where one section's error blocks saving another.
- *Error:* e.g., regenerating token fails → inline error, old token remains valid until regeneration succeeds.

## 8. Backup & Export — `/export` [Admin]

**Elements:** "Download database (.db)" button, "Download CSV export" button (contributions and expenses, either combined or as two files — see Assumptions), last-backup timestamp (from Litestream status, if surfaced — see `docs/TRD.md`), a short note (see below) explaining what the export does and does not contain.

**What's included / excluded:** both the `.db` download and the CSV export **exclude the `sessions` table and the `admins.password_hash` column** (see `docs/SCHEMA.md` §7 for the exact mechanism). The `.db` file an admin downloads here is a sanitized snapshot, not a raw copy of the live file — it contains members, contributions, expenses, group settings, and admin display names, but never session tokens/hashes or password hashes. This is called out on-screen (a small line under the download buttons: "This export excludes login credentials and session data — safe to share for backup or migration purposes.") so admins understand the export is safe to hand to a new host or another admin without also handing over a way to log in as someone else.

**States:**
- *Loading:* download buttons show a spinner while the file is prepared (for CSV, generated on demand; for `.db`, a sanitized consistent-snapshot copy — with `sessions` dropped and `password_hash` redacted — is streamed, never the live file handle).
- *Success:* browser's native file-download flow triggers; no in-app confirmation needed beyond the download itself.
- *Error:* "Couldn't generate the export. [Retry]" if the snapshot/export fails.

This screen exists independently of Litestream's automatic replication (`docs/TRD.md`), which is a separate, non-sanitized mechanism used only for disaster recovery (full state including sessions/admins, kept in a private bucket admins control) — see `docs/TRD.md` §9 and `docs/SCHEMA.md` §7 for why the two are allowed to differ. This on-demand export exists specifically so admins always have a manual, shareable copy — a product requirement from `docs/PRD.md` §5, not just a technical backup mechanism.

## 9. Public Transparency Page — `/p/:token` [Public]

**Purpose:** the trust-building public face of the app; no login.

**Layout:** group name, current balance, this-month collection/expense totals, recent activity feed (expenses always shown with description+amount+date; contributions shown with amount+date and, only if the privacy setting is ON, the member name), and, if the admin has generated one, the latest monthly summary text. If the operator has set a Privacy Policy URL (§7), a small footer link to it appears at the bottom of this page — omitted entirely if not set.

**States:**
- *Invalid/regenerated token:* generic 404-style page — "This page isn't available." No hint about whether the group exists at all, to avoid leaking information via error-message differences.
- *Empty (new group, no activity yet):* "This group hasn't recorded any activity yet."
- *Loading:* skeleton, same pattern as Dashboard.
- *Success:* populated read-only view. No edit affordances of any kind are rendered for public viewers — this is enforced server-side (the handler serving `/p/:token` never checks a session and never renders admin-only fragments), not just hidden via CSS.

## 10. Cross-Cutting Behaviors

- **Navigation:** all admin screens share a persistent bottom nav bar (mobile) with Dashboard, Add Contribution (center, emphasized), Members, Expenses, More (Summary/Settings/Export tucked behind this on the smallest screens — see `docs/UI_UX_BRIEF.md` for exact icon/label layout).
- **Session expiry:** any admin action attempted after session expiry redirects to `/login` with a notice: "Your session expired — please log in again." Unsaved form data is not preserved across this redirect in v1 (see Assumptions).
- **Network failure (HTMX request fails client-side, e.g., phone loses signal mid-save):** the affected fragment shows an inline error with a Retry action; the rest of the page remains interactive. This matters specifically because the primary device is a phone, where flaky connectivity is expected.
- **Confirmation dialogs** are used before any destructive or hard-to-reverse action: deactivating a member, regenerating the public token, removing an admin. Deleting a contribution/expense outright (vs. correcting via a new entry) requires a confirmation dialog and is itself an audited action (see `docs/SCHEMA.md`).
- **Language switch:** available from every screen (icon in header/nav), applies immediately without losing in-progress form data where technically feasible; RTL layout automatically applies for Arabic/Urdu (see `docs/UI_UX_BRIEF.md`).

## Assumptions

- CSV export produces two files (contributions.csv, expenses.csv) rather than one combined file, since the two record types have different columns — an admin wanting a combined view can do that in a spreadsheet themselves.
- Session-expiry during form entry does not attempt to preserve draft form data (e.g., via localStorage) in v1 — this is a small added complexity that can be revisited if it proves to be a frequent frustration in practice.
- Deleting a contribution/expense is a soft-delete at the data layer (flagged, not removed) so the audit trail remains intact even for deleted entries — detailed in `docs/SCHEMA.md`. The UI presents this to the admin as a normal "Delete," not "Soft delete," since the distinction is an implementation detail.
- The "More" overflow menu grouping (Summary/Settings/Export) on the smallest phone screens is a UX judgment call to keep the bottom nav to 4–5 icons max; exact grouping is finalized visually in `docs/UI_UX_BRIEF.md`.
