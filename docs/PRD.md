# Product Requirements Document — Sadqa Ledger

## 1. Overview

Sadqa Ledger is a small, self-hostable ledger web app for community groups — masjids, community associations, or any group that collects recurring member contributions and spends them on shared upkeep. It replaces a paper notebook with a mobile-first web app that any group can run on its own, at zero infrastructure cost.

"Sadqa" (also spelled "sadaqah") refers to voluntary charitable giving in Islamic practice. The name reflects the app's origin, but the app itself is generic: any group collecting dues and tracking shared expenses can use it (a residents' welfare association, a school parent-committee fund, a sports club). Nothing in the data model or UI assumes a specific religion, beyond bilingual labels that include Hindi terms alongside English (see `docs/UI_UX_BRIEF.md`).

## 2. Problem Statement

A volunteer (in our case, a person named Sohail) manually records every member's monthly contribution in a paper notebook: names in one column, amounts in the next. Each month he sums total collection by hand, subtracts expenses, and carries forward a running balance. Payment proof is shared informally as photos in a WhatsApp group.

This causes four concrete problems:

1. **Arithmetic errors.** Manual addition/subtraction over months of entries is error-prone, and errors compound into the running balance.
2. **Single point of failure.** The notebook is a physical object. If lost, damaged, or the volunteer is unavailable, the group loses its entire financial history.
3. **No shared visibility.** Members can't easily check whether they've paid this month, or see the group's overall financial position, without asking the volunteer directly.
4. **No audit trail.** There's no record of who entered or changed a number, which matters because this is donated trust money (amanah) — accuracy and transparency are not optional.

## 3. Why It Matters

This is community trust money. People contribute expecting it to be tracked accurately and spent on what it was collected for. A tool that mismanages this, or that is opaque about balances, undermines the trust the whole system depends on. The application's non-negotiable requirements are therefore **accuracy, transparency, and data durability** — not speed of feature delivery or visual polish.

## 4. Target Users

| Role | Who | Access |
|---|---|---|
| Admin | 2–3 trusted people per group (e.g., Sohail and one or two others) who physically collect money or receive payment confirmations | Full read/write: members, contributions, expenses, settings, exports |
| Public viewer | Any community member or outsider with the public link | Read-only, via unguessable token URL; sees what the group's privacy setting allows |

There is no per-member login in v1. Members are records in the ledger, not user accounts — see §8, Excluded Features.

## 5. Core Features (MVP)

1. **Members** — create, edit, deactivate members. Each member has a name and active/inactive status. Inactive members are excluded from "who hasn't paid" checklists but their historical contributions remain intact.
2. **Contributions** — admin logs a payment: select member (searchable list), enter amount, select month. Quick-amount chips (e.g., ₹200 / ₹500 / ₹1000) speed up entry for common amounts. One contribution record per member per month is the expected pattern (see `docs/SCHEMA.md` for how partial/multiple payments in a month are handled).
3. **Expenses** — admin logs an expense: description, amount, date, optional receipt photo.
4. **Dashboard** — this month's total collection, all-time total collected, all-time total spent, current balance, and a paid/unpaid checklist of active members for the current month.
5. **Public transparency page** — a read-only page at an unguessable token URL (no login) showing collection and expense history. Respects the group's name-privacy setting (§6).
6. **Monthly summary for WhatsApp** — a formatted, copyable/shareable text summary of the month (total collected, total spent, balance, and — if the privacy setting allows — who paid) designed to be pasted directly into a WhatsApp group.
7. **Backup & export** — the raw SQLite database file and a CSV export of contributions/expenses are downloadable by admins at any time, independent of the automatic Litestream replication described in `docs/TRD.md`.

Every write (create/edit/delete of a member, contribution, or expense) records which admin performed it and when — the audit trail.

## 6. Privacy Setting

A single group-level setting: **"Show contributor names publicly."**

- **Default: OFF.** The public page and WhatsApp summary show total collected and total spent, but not who gave what. This lets people give quietly, which is a value many contributors hold.
- **When ON:** the public page and WhatsApp summary list member names alongside amounts.

Admins always see full detail regardless of this setting. This setting has no per-member override in v1 — it is all-or-nothing for the group (see Excluded Features).

## 7. User Stories

- As an admin, I can log a member's monthly contribution in under 15 seconds from my phone, faster than writing it in a notebook.
- As an admin, I can see at a glance which active members have not yet paid this month.
- As an admin, I can record an expense with a photo of the receipt so there's proof of what the money was spent on.
- As an admin, I can generate a monthly summary and paste it into our WhatsApp group without retyping numbers.
- As an admin, I can download the full database or a CSV at any time, so our data is never locked into one machine or one person.
- As a community member without a login, I can open a link and see the group's current balance and recent activity, without seeing individual names if the group has chosen to keep them private.
- As a group considering this software, I can look at the audit trail and know exactly who recorded each entry, so I can trust the numbers.

## 8. Success Metrics

1. **Entry speed**: logging one contribution takes an admin less time on a phone than writing the same entry in a paper notebook. This is the app's primary success metric — if this fails, the app has failed its core purpose.
2. **Zero data loss** over a 12-month period of normal use, verified by successful restore from Litestream backup at least once.
3. **Adoption beyond the origin group**: at least one other community forks and self-hosts the repo within the first year.
4. **Zero recurring infrastructure cost** for a typical single-masjid deployment (see `docs/TRD.md` §Deployment).

## 9. MVP Scope Boundary

In scope for v1: everything in §5.

Explicitly out of scope for v1 (see §10 for the full excluded list). If a request doesn't clearly serve accuracy, transparency, durability, or mobile entry speed, it is deferred past v1.

## 10. Explicitly Excluded from v1

- **Per-member logins / accounts.** Members are data, not users. Only admins authenticate. Revisit only if a group specifically needs members to self-report.
- **Multi-tenancy.** One deployed instance serves exactly one masjid/group. Other groups get their own instance by forking the repo and self-hosting (see `docs/TRD.md` §Architecture). There is no central hosted service and no cross-tenant data model.
- **Payment processing / online payments.** The app records that a payment happened; it does not move money. Contributions are collected in person or via existing channels (cash, UPI, bank transfer) and simply logged afterward.
- **Per-member privacy overrides.** The name-privacy setting is group-wide only.
- **Recurring/auto-billing or payment reminders.** No automated notifications to members in v1 (a manual WhatsApp share covers this need for now).
- **Role granularity beyond admin/public.** No "viewer with login," no "treasurer vs. secretary" permission tiers in v1 — all admins have equal, full access.
- **Multi-currency support.** Single currency per deployment, configured at setup (see `docs/TRD.md`); amounts are stored as integers in minor units to avoid floating-point error, but there is no in-app currency conversion.
- **Native mobile apps.** The PWA (installable, home-screen-capable web app) covers the "feels like an app" requirement without maintaining separate iOS/Android codebases.
- **Advanced reporting/analytics** (charts, year-over-year trends, forecasting) beyond the dashboard totals and CSV export. CSV export lets any admin do deeper analysis in a spreadsheet if they want it.

## Assumptions

- "Community group" in this document defaults to a masjid for concreteness, per the origin story, but no feature depends on that being true.
- Currency symbol/code is a per-deployment configuration value (e.g., ₹ for INR), not hardcoded, so other communities can self-host in their own currency — see `docs/TRD.md` and `docs/SCHEMA.md` for where this is configured.
- "Unguessable token URL" means a long, random, unindexed path segment (e.g., a UUID or equivalent), not a password. It provides obscurity, not authentication. This tradeoff is intentional for v1 given the "near-zero technical skill" deployment goal — see `docs/TRD.md` §Security for the full reasoning and threat model.
- One contribution per member per month is the expected common case; the schema permits multiple contribution records per member per month (e.g., a partial payment followed by a top-up) rather than forcing edits to a single row — see `docs/SCHEMA.md`.
