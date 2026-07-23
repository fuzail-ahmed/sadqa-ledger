# Sadqa Ledger

A tiny, free-to-run ledger app for community groups — masjids, resident associations, or any group — that collect monthly contributions and spend them on shared upkeep (repairs, electricity, fans, doors, and so on).

## Why this exists

In our town, a volunteer named Sohail records every member's monthly contribution by hand in a paper notebook, then manually adds it all up and subtracts expenses to know the group's balance. Manual arithmetic causes mistakes, the notebook can be lost, and nobody else can easily see who has paid this month.

Sadqa Ledger replaces the notebook with a simple phone-friendly app that any 2–3 trusted people can use to log contributions and expenses in seconds, while giving the whole community an always-up-to-date, transparent view of the group's finances — without needing everyone to have a login.

This is trust money (amanah), donated in good faith. Accuracy, transparency, and never losing the data are the whole point of this project — not flashy features.

Full reasoning behind every product and technical decision lives in [`docs/`](docs/):

- [`docs/PRD.md`](docs/PRD.md) — what this app is and isn't
- [`docs/TRD.md`](docs/TRD.md) — the tech stack and why
- [`docs/APP_FLOW.md`](docs/APP_FLOW.md) — every screen, in detail
- [`docs/UI_UX_BRIEF.md`](docs/UI_UX_BRIEF.md) — visual/design direction
- [`docs/SCHEMA.md`](docs/SCHEMA.md) — the database, in detail
- [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md) — how it's being built
- [`docs/DEPLOY.md`](docs/DEPLOY.md) — how to run your own instance

## Screenshots

*(Coming soon — screenshots of the Dashboard, Add Contribution, and Public Transparency Page will go here once the first version is built.)*

## Features

- **Log a contribution in seconds** — pick a member, tap an amount (or type your own), save. Built to be faster than writing it in a notebook.
- **Dashboard** — this month's collection, all-time totals, current balance, and a paid/unpaid checklist for the current month, all at a glance.
- **Expense tracking** — record what was spent, with an optional photo of the receipt.
- **Public transparency page** — a shareable link (no login needed) where anyone in the community can see the group's balance and activity. Whether member names are shown publicly is a setting your group controls — hidden by default, so people can give quietly if they prefer.
- **Monthly WhatsApp summary** — a ready-to-paste text summary of the month's activity for your group's WhatsApp.
- **Backup & export, always available** — download the full database or a CSV export any time, so your group's data is never locked to one device or one person.
- **Every entry is attributed** — the app records which admin logged each contribution or expense, so the numbers are always accountable.
- **Works in English, Hindi, and Arabic.**
- **Installs like an app** on your phone's home screen — no app store needed.

## Quick start (Docker Compose)

The production configuration pulls prebuilt, multi-architecture Docker images from GitHub Container Registry (GHCR) by default. These images are automatically built and published by GitHub Actions on every release.

If you already have Docker installed and want to run the prebuilt image:

```bash
cp .env.example .env
# edit .env: set DOMAIN, BASE_URL, SESSION_SECRET, and backup settings
docker compose pull
docker compose up -d
```

### Local Development

If you are developing locally and want to build the Docker image from your local source files:

```bash
cp .env.example .env
cp docker-compose.override.yml.example docker-compose.override.yml
# edit .env as needed
docker compose build
docker compose up -d
```
### Image Versioning and Hardening

Production deployments parameterize the application version using the `APP_VERSION` environment variable in the `.env` file. This offers several production advantages:
- **Versioning Control:** You can pin deployments to specific versions instead of always tracking `latest`.
- **Easy Rollbacks:** If a new update introduces bugs, rolling back is as simple as editing `APP_VERSION` in your `.env` file to a previous stable tag and running `docker compose up -d`. You never need to edit the `docker-compose.yml` file.

Available version patterns for `APP_VERSION`:
- `latest` (default): Tracks the newest build of the `master`/`main` branch.
- **Release Tags** (e.g., `v1.0.3`): Pins the application to a specific tagged release.
- **Git SHAs** (e.g., `3d9e81f`): Pins the application to a specific commit build for precise, reproducible deployment control.

Then open the `BASE_URL` you configured and follow the first-run setup to create your first admin account and configure your group.

Don't have a `.env` file yet? Copy [`.env.example`](.env.example) to `.env` and fill in your own values first.

This runs the app for trying it out. **To actually put this online for your community — on a free server, with a real domain and automatic HTTPS, and continuous off-site backups — follow the comprehensive [`docs/DEPLOY.md`](docs/DEPLOY.md) guide.** It's written for someone who has never seriously used a terminal before.

## Tech stack

Go, `chi`, `templ`, HTMX, Tailwind CSS (no Node.js build step), SQLite (WAL mode), Litestream backups to Cloudflare R2, session-cookie auth, a single Docker image, deployed to a free Oracle Cloud VM behind Caddy. Full reasoning for every choice is in [`docs/TRD.md`](docs/TRD.md). Total infrastructure cost target: **$0/month**.

## Self-hosting your own instance

Sadqa Ledger is built for **one instance per community** — there's no shared multi-tenant hosted version. If your masjid or group wants to use this, you fork this repository and run your own copy, so your group keeps full custody of its own data. Start with [`docs/DEPLOY.md`](docs/DEPLOY.md).

## Contributing

Contributions are welcome — code, translations, documentation, or bug reports. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for local setup, code style, and how to add a new language. Please also read [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md). If you've found a security issue, see [`SECURITY.md`](SECURITY.md) instead of opening a public issue.

## Legal

Sadqa Ledger is self-hosted, open-source software — we (the authors) never see or store any operator's data; each group that runs an instance is responsible for its own members' data. See:

- [`legal/DISCLAIMER.md`](legal/DISCLAIMER.md) — our position as software authors: no warranty, not a data controller, not accounting software.
- [`legal/PRIVACY_POLICY_TEMPLATE.md`](legal/PRIVACY_POLICY_TEMPLATE.md) and [`legal/TERMS_TEMPLATE.md`](legal/TERMS_TEMPLATE.md) — fill-in-the-blank documents for operators to publish for their own members.
- [`docs/OPERATOR_RESPONSIBILITIES.md`](docs/OPERATOR_RESPONSIBILITIES.md) — a plain-language checklist for whoever deploys and runs an instance.
- [`docs/ACCESSIBILITY.md`](docs/ACCESSIBILITY.md) — our accessibility commitments (WCAG 2.1 AA) and how we test them.

## License

MIT — see [`LICENSE`](LICENSE).

## A note on how this was built

Parts of this project — including planning documents, templates, and code — were written with AI assistance. Everything is reviewed by a human before being merged; if you spot something that looks off, please open an issue.

---

## हिन्दी में संक्षिप्त जानकारी

Sadqa Ledger एक छोटा, मुफ़्त चलने वाला ऐप है जो मस्जिद या किसी भी सामुदायिक समूह के लिए बनाया गया है, जो हर महीने चंदा इकट्ठा करते हैं और उसे मरम्मत, बिजली बिल जैसे खर्चों पर लगाते हैं। यह कागज़ की नोटबुक की जगह लेता है — कोई भी 2-3 भरोसेमंद व्यक्ति फ़ोन से कुछ ही सेकंड में चंदा दर्ज कर सकते हैं, और पूरा समुदाय एक लिंक के ज़रिए हिसाब-किताब पारदर्शी रूप से देख सकता है। सदस्यों के नाम सार्वजनिक रूप से दिखाना या न दिखाना, यह आपके समूह की पसंद पर निर्भर करता है। पूरी जानकारी के लिए [`docs/DEPLOY.md`](docs/DEPLOY.md) देखें।
