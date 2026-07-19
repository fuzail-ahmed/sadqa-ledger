# Contributing to Sadqa Ledger

Thanks for considering a contribution. This project is small on purpose — see [`docs/TRD.md`](docs/TRD.md) for the reasoning behind the tech choices before proposing a new dependency or framework.

## Local development setup

Prerequisites: Go (latest stable), the [templ CLI](https://templ.guide), and the [Tailwind standalone CLI](https://tailwindcss.com/blog/standalone-cli) (no Node.js needed for either).

```bash
git clone https://github.com/fuzail-ahmed/sadqa-ledger.git
cd sadqa-ledger
cp .env.example .env          # then edit .env with local values

# Generate templ files
templ generate

# Build the Tailwind CSS
./tailwindcss -i web/static/css/input.css -o web/static/css/output.css --watch

# In a separate terminal, run the app
go run ./cmd/server
```

Open `http://localhost:8080`. The first run will walk you through creating an admin account and a group name — see [`docs/APP_FLOW.md`](docs/APP_FLOW.md) §First-run setup.

Run tests:

```bash
go test ./...
```

Run migrations manually if needed (the app also applies them automatically on startup, see [`docs/SCHEMA.md`](docs/SCHEMA.md) §8):

```bash
go run ./cmd/migrate
```

## Code style

- Standard Go formatting: run `gofmt`/`go vet` before committing (CI will reject unformatted code).
- No ORM, no query builder — write plain SQL with `database/sql` placeholders (see [`docs/TRD.md`](docs/TRD.md) §5). Never build SQL with string concatenation.
- Keep templates (`templ`) free of business logic — compute values in Go, pass them in as simple template data.
- Use descriptive names over abbreviations (`contributionMonth`, not `cm`).
- Money is always an integer in minor units (paise, cents) — never a float. See [`docs/SCHEMA.md`](docs/SCHEMA.md) §1.
- New UI strings go into the JSON translation files (`i18n/en.json`, etc.) with a descriptive key — never hardcode user-facing text in a template.

## Commit conventions

Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`. Example: `feat: add quick-amount chips to contribution form`. Keep commits focused — one logical change per commit.

## Pull requests

- Reference the relevant doc/section your change implements (e.g., "implements Add Contribution per `docs/APP_FLOW.md` §3").
- Include the states you tested (empty/loading/error/success — see [`docs/APP_FLOW.md`](docs/APP_FLOW.md)) in the PR description.
- Keep PRs small and reviewable; a phase from [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md) may span multiple PRs.
- CI (build, `go vet`, tests) must pass before merge.
- See [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) for the checklist that will pre-fill your PR description.

## Adding a new language

1. Copy `i18n/en.json` to `i18n/<language-code>.json` (use the [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639_language_codes) two-letter code, e.g. `bn` for Bengali).
2. Translate every value, keeping the keys unchanged.
3. If the language is right-to-left (like Arabic and Urdu already are), confirm the layout mirrors correctly — see [`docs/UI_UX_BRIEF.md`](docs/UI_UX_BRIEF.md) §8 for the RTL approach (CSS logical properties, not hardcoded left/right).
4. Add the language to the switcher's language list and to `docs/TRD.md`'s supported-languages list if you'd like it documented as officially supported.
5. Open a PR — translations are one of the easiest and most valued ways to contribute, even without writing any Go.

## Reporting bugs / requesting features

Use the templates under [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/). For security vulnerabilities, do **not** open a public issue — see [`SECURITY.md`](SECURITY.md).
