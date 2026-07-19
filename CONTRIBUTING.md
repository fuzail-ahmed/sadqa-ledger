# Contributing to Sadqa Ledger

Thanks for considering a contribution. This project is small on purpose — see [`docs/TRD.md`](docs/TRD.md) for the reasoning behind the tech choices before proposing a new dependency or framework.

## Local development setup

Prerequisites: **Go** and **make**. That's it — you do not need to install templ, the Tailwind CLI, or Node.js yourself; `make setup` handles all of that for you.

```bash
git clone https://github.com/fuzail-ahmed/sadqa-ledger.git
cd sadqa-ledger
make setup
make dev
```

Then open `http://localhost:8080`. The first run will walk you through creating an admin account and a group name — see [`docs/APP_FLOW.md`](docs/APP_FLOW.md) §First-run setup.

`make setup` verifies your Go version, installs `templ` and `air` (pinned versions, via `go install`), downloads the Tailwind standalone CLI **v4** for your OS/architecture into a local `.tools/` folder, downloads Basecoat's CSS/JS bundle into `web/static/vendor/basecoat/` (vendored, not npm-installed — see [`docs/TRD.md`](docs/TRD.md) §4), creates your `.env` from `.env.example` (never overwriting one that already exists), downloads Go module dependencies, and builds once so you know it all works. It's safe to run more than once.

`make dev` runs the app with hot reload via [air](https://github.com/air-verse/air): saving a `.go`, `.templ`, `.css`, or i18n `.json` file regenerates templ, rebuilds the CSS, and rebuilds/restarts the server automatically. One terminal, no watch script in a second window.

### Windows

`make` isn't available in a plain Windows Command Prompt or PowerShell by default. Use one of these instead (all work fine with the Makefile as-is — no separate scripts needed):

- **Git Bash** (installed alongside [Git for Windows](https://git-scm.com/download/win)) — includes `make`-compatible tools for everything this Makefile needs (`curl`, `awk`, `sed`, `uname`). This is the quickest option if you already have Git for Windows.
- **WSL** (Windows Subsystem for Linux) — run everything inside a real Linux environment; the most trouble-free option if you're doing more than occasional contributions.
- **Chocolatey**: `choco install make`, then run the same commands from PowerShell or Command Prompt.

Whichever you pick, `make setup` and `make dev` are the same two commands — the Makefile detects Windows automatically and downloads the right Tailwind CLI binary for it.

### Other useful commands

Run `make help` at any time to see the full list with descriptions. The ones you'll reach for most:

```bash
make build    # production binary in bin/, minified CSS
make test     # go test ./...
make lint     # gofmt + go vet
make fmt      # gofmt -w .
make templ    # regenerate .templ files only
make css      # rebuild Tailwind CSS only
make migrate  # run pending DB migrations manually (the app also applies them on startup — see docs/SCHEMA.md §8)
make clean    # remove build artifacts (never touches .env or your database)
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
3. If the language is right-to-left (like Arabic already is), confirm the layout mirrors correctly — see [`docs/UI_UX_BRIEF.md`](docs/UI_UX_BRIEF.md) §8 for the RTL approach (CSS logical properties, not hardcoded left/right).
4. Add the language to the switcher's language list and to `docs/TRD.md`'s supported-languages list if you'd like it documented as officially supported.
5. Open a PR — translations are one of the easiest and most valued ways to contribute, even without writing any Go.

## Reporting bugs / requesting features

Use the templates under [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/). For security vulnerabilities, do **not** open a public issue — see [`SECURITY.md`](SECURITY.md).
