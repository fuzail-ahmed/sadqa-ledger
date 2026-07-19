# Security Policy

## This software handles community financial records

Sadqa Ledger stores real, trust-based financial data belonging to community groups — member contributions, expenses, and (for self-hosted instances) admin login credentials. Please take security reports seriously and report privately rather than publicly, per the process below.

## Supported Versions

This project is pre-1.0 and moves quickly. Security fixes are applied to the latest released version only.

| Version | Supported |
|---|---|
| Latest release (`main` / most recent tag) | Yes |
| Older tagged releases | No — please upgrade |

Once the project reaches 1.0, this table will be updated to reflect a longer support window for stable major versions.

**A fix released here only protects your instance once you apply it.** Sadqa Ledger is self-hosted software — we cannot patch, restart, or update any operator's running instance ourselves, because we have no access to it (see [`legal/DISCLAIMER.md`](legal/DISCLAIMER.md)). When a security fix is released, it is each operator's responsibility to pull the new image and redeploy, per the upgrade steps in [`docs/DEPLOY.md`](docs/DEPLOY.md) §Upgrading. Watch the repository (GitHub's "Watch" → "Custom" → "Releases") if you want to be notified when a new version is published.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Instead, report privately using one of these methods:

1. **GitHub private vulnerability reporting** (preferred): go to the repository's **Security** tab → **Report a vulnerability**. This opens a private advisory visible only to maintainers.
2. **Email**: if private reporting is unavailable, email the maintainer listed in the repository's GitHub profile with a clear subject line (e.g., "SECURITY: Sadqa Ledger — [brief description]").

Please include:

- A description of the vulnerability and its potential impact.
- Steps to reproduce, or a proof of concept if possible.
- The version/commit you tested against.

## What to expect

- Acknowledgement of your report within a reasonable time.
- An assessment of severity and, if confirmed, a plan and rough timeline for a fix.
- Credit in the release notes/changelog once fixed, unless you prefer to remain anonymous.

## Scope

In scope: the application code in this repository (authentication, session handling, data access, file uploads, the public transparency page, exports/backups).

Out of scope: vulnerabilities in third-party infrastructure you choose for self-hosting (your Oracle Cloud VM configuration, your Cloudflare R2 account security, your Caddy/TLS setup) — though `docs/DEPLOY.md` issues affecting the documented setup steps are welcome as regular bug reports or security reports as appropriate.

## Notes on this project's security posture

For transparency, some relevant design decisions and their rationale are documented rather than hidden:

- Session tokens are hashed (SHA-256) before storage — see [`docs/SCHEMA.md`](docs/SCHEMA.md) §5 and [`docs/TRD.md`](docs/TRD.md) §6.
- Passwords are hashed with bcrypt — see [`docs/TRD.md`](docs/TRD.md) §6.
- The public transparency page relies on an unguessable token URL rather than a login — this is a deliberate, documented tradeoff; see [`docs/TRD.md`](docs/TRD.md) §9 for the threat model.
- On-demand database/CSV exports exclude session data and password hashes by design — see [`docs/SCHEMA.md`](docs/SCHEMA.md) §7 and [`docs/APP_FLOW.md`](docs/APP_FLOW.md) §8.

If you believe any of these documented tradeoffs is unsound, that's a welcome and valid security report too.
