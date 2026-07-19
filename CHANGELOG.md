# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project planning documents: PRD, TRD, App Flow, UI/UX Brief, Schema, and Implementation Plan (`docs/`).
- Repository scaffolding: README, LICENSE (MIT), `.gitignore`, `.env.example`, contributing/security/conduct docs, issue and PR templates.

## [0.1.0] - 2026-07-19

### Added
- Core admin authentication and first-run setup wizard.
- Member management (creating, editing, and toggling active/inactive status).
- Contribution logging (with quick-amount chip triggers, member type-ahead search, and duplication warnings).
- Interactive Dashboard (with all-time and monthly totals, bilingual Hindi/English labels, and monthly checklist).
- Expense tracking (with receipt photo uploads, local file storage, and audit trail soft-deletes).
- Public transparency ledger page (`/p/:token`) with configurable name-privacy configurations and search crawler indexing protection.
- Admin settings dashboard (supporting privacy toggles, public link token regeneration, group info, admins list, and language selection).
- WhatsApp-shareable monthly text summary.
- Backup & export (on-demand sanitized SQLite database snapshot download and CSV exports for contributions and expenses).
- Multi-lingual support (complete English, Hindi, and Arabic translations, featuring automatic RTL document mirroring for Arabic).
- Progressive Web App (PWA) manifest registration, service worker offline shell caching, and home-screen installability.

[0.1.0]: https://github.com/fuzail-ahmed/sadqa-ledger/releases/tag/v0.1.0
