## What does this PR do?

Brief description of the change.

## Which doc/section does this implement?

Reference the relevant part of `docs/` (e.g., "Implements Add Contribution per `docs/APP_FLOW.md` §3" or "Phase 5 of `docs/IMPLEMENTATION_PLAN.md`").

## States tested

If this touches a UI screen, confirm which states from `docs/APP_FLOW.md` you verified:

- [ ] Empty state
- [ ] Loading state
- [ ] Error state
- [ ] Success state
- [ ] Validation errors (if applicable)

## Checklist

- [ ] `gofmt`/`go vet` pass locally
- [ ] `go test ./...` passes locally
- [ ] No string-concatenated SQL introduced (parameterized queries only — see `docs/TRD.md` §9)
- [ ] No new user-facing text hardcoded outside the i18n JSON files
- [ ] Money values handled as integers in minor units, not floats (see `docs/SCHEMA.md` §1)
- [ ] Updated relevant docs if this PR changes behavior described in `docs/`
- [ ] Added a `CHANGELOG.md` entry under `[Unreleased]` if user-facing

## Screenshots (if UI change)

Before/after screenshots, ideally on a phone-width viewport per the mobile-first requirement.

## Anything else reviewers should know?
