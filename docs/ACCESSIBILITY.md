> This document is not legal advice on accessibility law (e.g., ADA, EN 301 549, the Rights of Persons with Disabilities Act in India) — it states our design/engineering commitments and how we verify them. If your organization has specific legal accessibility obligations, confirm compliance with a qualified professional.

# Accessibility — Sadqa Ledger

Sadqa Ledger is used by people who are often not digital natives — community elders, volunteers with no technical background, people with aging eyesight or motor control. Accessibility here isn't a checkbox added at the end; it's the same audience described throughout `docs/UI_UX_BRIEF.md` §1 ("strong readability for older users") and `docs/PRD.md` (Sohail's phone as the primary device). This document makes those commitments concrete and testable.

## Target standard

**WCAG 2.1, Level AA**, as the baseline for every screen in `docs/APP_FLOW.md`, including the public transparency page (which has no login, so it must work for the widest possible range of visitors).

## Commitments

### Color and contrast

- **Minimum 4.5:1 contrast ratio** for normal text, and 3:1 for large text (18pt+/14pt+ bold) and meaningful UI graphics (icons that convey state, like the paid/unpaid checkmark), against their background.
- Color is never the only signal for meaning: the paid/unpaid badges in `docs/UI_UX_BRIEF.md` §4 already pair color with an icon (checkmark vs. dash) for this reason; the same rule applies to any new status indicator added later (e.g., a red/green balance color must be paired with a +/− sign or label, not color alone).
- The palette in `docs/UI_UX_BRIEF.md` §2 is documented as a starting point pending a real contrast check — see that document's Assumptions section. This document formalizes that check as a required verification step (see Testing Checklist below), not an optional nice-to-have.

### Touch and interaction targets

- **Minimum 44×44px** for every tappable element — buttons, checklist rows, nav icons, quick-amount chips — already specified in `docs/UI_UX_BRIEF.md` §6. This document reaffirms it as an accessibility requirement, not just a mobile-usability one: it also serves users with tremor or reduced fine motor control.
- Adequate spacing between adjacent tap targets (at least 8px) so an imprecise tap doesn't land on the wrong element — for example, adjacent Edit/Deactivate buttons on a member row (`docs/APP_FLOW.md` §4).

### Keyboard navigation

- **Every action must be reachable and completable using only a keyboard** — tab order follows visual/logical order, all interactive elements (buttons, form fields, the member search/type-ahead in `docs/APP_FLOW.md` §3, HTMX-swapped content) are focusable and operable with Enter/Space, and no functionality depends on a mouse-only interaction (hover-to-reveal is already excluded by `docs/UI_UX_BRIEF.md` §6's "no hover-dependent interactions" rule, which also serves keyboard-only and touch users).
- Modal dialogs and confirmation dialogs (deactivate member, regenerate token, delete a contribution — `docs/APP_FLOW.md` §10) trap focus while open and return focus to the triggering element on close.

### Visible focus indicators

- Every focusable element shows a clearly visible focus outline (not `outline: none` without a replacement) that meets the same 3:1 contrast minimum against its background. This matters as much as it does for touch, because an admin correcting an entry from a desktop browser, or any keyboard user, needs to see where they are on the page.

### Screen readers and semantic markup

- Proper semantic HTML first (`<button>`, `<label>`, `<nav>`, headings in order) rather than `<div>`-with-click-handler patterns — this is largely free given `templ`'s plain-HTML output (`docs/TRD.md` §4) and just requires discipline in how templates are written.
- **ARIA labels** where semantic HTML alone isn't enough to convey meaning — e.g., an icon-only nav button (`docs/UI_UX_BRIEF.md` §7) needs an `aria-label` describing its action, not just a visual icon.
- HTMX fragment swaps (`docs/TRD.md` §4) that update page content out of the user's current focus (e.g., the Dashboard checklist refreshing after a save) use an ARIA live region so screen reader users are told the update happened, not left with silently stale-sounding content.
- Form validation errors (`docs/APP_FLOW.md` §3, §5) are programmatically associated with their field (`aria-describedby` or equivalent), not just visually adjacent, so a screen reader announces the specific error when the field is focused.

### Images and receipt photos

- **All images have alt text.** For an expense's optional receipt photo (`docs/SCHEMA.md` §3.4, `docs/APP_FLOW.md` §5), the alt text is generated from the expense description and date (e.g., "Receipt photo for: Electricity bill, 12 July 2026") rather than left blank or generic ("image1.jpg"), so a screen reader user gets equivalent context to a sighted user glancing at the thumbnail.
- Decorative icons that duplicate an adjacent text label (e.g., a checkmark icon next to the word "Paid") are marked as decorative (`aria-hidden="true"`) so they aren't announced redundantly.

### Motion

- The app respects the operating system's `prefers-reduced-motion` setting: any transition or animation (toast fade-in, HTMX swap transition) is reduced to an instant or near-instant change for users who've requested less motion, consistent with `docs/UI_UX_BRIEF.md` §9's existing "confirm, don't celebrate" principle — there was never meant to be flashy animation here, but this makes the reduced-motion case an explicit, tested requirement rather than an accidental side effect of a calm design.

### Font size and readability

- **16px minimum body text**, generous line height (1.5+), and no thin/light font weights — already specified in `docs/UI_UX_BRIEF.md` §3. This document adds: text must be resizable up to 200% via the browser's zoom without loss of content or functionality (no fixed-height containers that clip enlarged text), since some older users rely on browser zoom rather than an in-app font-size setting.

## Testing checklist

Run this checklist before shipping any new screen, and again as part of Phase 10 (`docs/IMPLEMENTATION_PLAN.md`, Testing & Hardening):

1. **Automated scan — Lighthouse.** Run Chrome DevTools' Lighthouse Accessibility audit on every screen in `docs/APP_FLOW.md` §0's route map. Target score: 100, or every sub-100 finding explicitly triaged (fixed, or logged with a reason it doesn't apply).
2. **Automated scan — axe.** Run the axe browser extension (or `@axe-core/cli` in CI) as a second automated pass — axe and Lighthouse catch overlapping but not identical issues, so both are used rather than relying on one.
3. **Keyboard-only pass.** Unplug the mouse. Complete the full core flow — log in, add a contribution, add an expense, view the dashboard, log out — using only Tab, Shift+Tab, Enter, Space, and arrow keys. Confirm nothing is unreachable and focus order makes sense.
4. **Screen reader spot check.** Using a free screen reader (NVDA on Windows, VoiceOver on Mac/iOS, TalkBack on Android — pick at least one, ideally the one matching your test device), walk through the Add Contribution flow and the public transparency page and confirm labels, errors, and live-region updates are announced sensibly.
5. **Contrast check.** Verify the actual implemented colors (not just the documented hex values in `docs/UI_UX_BRIEF.md` §2) against a contrast checker (e.g., the one built into Chrome DevTools' color picker) for text, icons, and focus indicators.
6. **Zoom check.** Set browser zoom to 200% on a phone-width viewport and confirm no content is clipped or overlapping, and the Add Contribution form's Save button is still reachable.
7. **Reduced motion check.** Enable "reduce motion" in the OS accessibility settings and confirm transitions/animations are minimized or removed.

Where this checklist finds a gap in `docs/UI_UX_BRIEF.md`, fix the UI brief too, not just the code — see the note added there (§10, new) as part of this document's creation.

## Where this fits in the build plan

Accessibility acceptance criteria have been added to the relevant phases in `docs/IMPLEMENTATION_PLAN.md` (Phase 3 Core Admin UI Shell, Phase 9 i18n & PWA Completion, and Phase 10 Testing & Hardening) rather than treated as a single late-stage task — catching a missing alt-text pattern or bad tab order early, when only a few screens exist, is far cheaper than retrofitting it across a finished app.

## Assumptions

- WCAG 2.1 AA (not AAA) is chosen as the target, matching common practice for this kind of civic/community software and matching the contrast ratios already implied by `docs/UI_UX_BRIEF.md`'s "older-user audience" framing; AAA's stricter 7:1 contrast and other requirements are not adopted as a strict requirement but are not discouraged if easy to achieve.
- Automated tooling (Lighthouse, axe) is treated as a floor, not a substitute for the manual keyboard/screen-reader passes — automated scanners catch roughly a third to half of real accessibility issues in practice, hence both an automated and manual step are required in the testing checklist.
- Receipt photo alt text is auto-generated from existing expense fields (description + date) rather than requiring the admin to type separate alt text, to avoid adding friction to the expense-entry flow described in `docs/APP_FLOW.md` §5 — this keeps accessibility from competing with the "fast entry" product goal.
