# UI/UX Brief — Sadqa Ledger

Design direction for Sadqa Ledger. Screens and behaviors are specified in `docs/APP_FLOW.md`; this document covers the visual and interaction language that ties those screens together. Product context is in `docs/PRD.md`.

## 1. Design Character

**Calm and trustworthy, not flashy.** This app displays other people's donated money. The design should feel like a well-kept ledger book, not a fintech startup pitch deck: quiet colors, clear numbers, generous whitespace, no marketing gradients, no celebratory animations on saving a contribution. Confidence comes from clarity and consistency, not decoration.

**Strong readability for older users.** Several likely users (community elders, volunteers like Sohail) are not digital natives and may have lower vision acuity. This drives concrete rules in §4 (type sizes) and §6 (touch targets) — these are not optional "nice to have" accessibility extras, they are core requirements given the actual user base.

**Bilingual by default.** Key financial terms appear in both English and the local-language term wherever they're a heading or label, not just in the language switcher. Example: a stat card reads "This Month / जमा" or "Balance / बाकी" rather than switching entirely to one language and hiding the other. This is a deliberate v1 decision — see §8.

## 2. Color Palette

A restrained palette; color is used to communicate meaning (paid/unpaid, income/expense), not for decoration.

| Token | Use | Value (approx.) |
|---|---|---|
| `--color-bg` | Page background | Off-white, `#FAF9F6` |
| `--color-surface` | Cards, panels | White, `#FFFFFF` |
| `--color-text-primary` | Body text, numbers | Near-black, `#1F2422` |
| `--color-text-muted` | Secondary labels, timestamps | Slate gray, `#6B7280` |
| `--color-primary` | Primary actions (Save, Add Contribution), links | Deep teal/green, `#0F6B5C` |
| `--color-primary-hover` | Hover/active state of primary | `#0B5548` |
| `--color-income` | Contributions, positive balance | Same deep green family, `#0F6B5C` |
| `--color-expense` | Expenses | Muted terracotta/brick, `#B4552F` |
| `--color-warning` | Unpaid indicator, soft warnings | Amber, `#B8860B` |
| `--color-danger` | Destructive actions, errors | Muted red, `#B3261E` |
| `--color-border` | Dividers, card borders | Light gray, `#E5E1DA` |

Green/teal was chosen over a generic blue because it reads as "growth/ledger/trust" without borrowing bank-app blue clichés, and because it pairs legibly with the terracotta expense color for colorblind-safe income/expense distinction (verify with a contrast/colorblindness check during implementation — see Assumptions). Colors are defined as CSS variables so Basecoat/Tailwind theming stays centralized (`docs/TRD.md` §Frontend).

## 3. Typography

- **Typeface:** a highly legible system-first stack (e.g., `system-ui`, falling back to Noto Sans for broad Devanagari/Arabic script coverage where the OS font lacks it). No custom display font — legibility across three languages and older eyes outranks brand personality here.
- **Base size:** 16px minimum for body text on mobile; numbers on the Dashboard stat cards are notably larger (24–32px) since they're the most-glanced-at content on the screen.
- **Line height:** generous (1.5+) for body text, especially important for Arabic script and for older readers.
- **Weight:** avoid thin/light weights entirely; primary content uses regular or medium weight, headings use semibold. Thin fonts fail the "readable for older users" requirement.
- **Numerals:** contribution/expense amounts always rendered in a tabular/monospaced numeral style so columns of numbers align — makes scanning a list of amounts easier, echoing the notebook's column alignment this app is replacing.

## 4. Component Style

Basecoat (shadcn-like components on plain HTML, per `docs/TRD.md`) gives the baseline shapes; house rules on top of it:

- **Cards:** subtle border (`--color-border`) plus a very light shadow, not heavy elevation — calm, not "app-store flashy."
- **Buttons:** primary action = solid `--color-primary` fill, full-width on mobile forms (Save, Add Contribution). Secondary/cancel actions = outline or ghost style, never competing visually with the primary action on the same screen.
- **Badges:** Active/Inactive member status, Paid/Unpaid checklist markers use small pill badges with color + icon (not color alone, for colorblind accessibility) — e.g., a checkmark icon for paid, a dash for unpaid, not just green/gray dots.
- **Forms:** labels always visible above the field (never placeholder-only labels, which disappear once typing starts and hurt usability for users who type slowly or get interrupted).
- **Toasts/confirmations:** small, bottom-anchored, auto-dismissing after a few seconds, non-blocking — confirm the action happened without demanding a dismiss tap.
- **Empty states:** a short sentence plus one clear call-to-action button, no cartoon illustrations — consistent with the calm, unflashy character.

## 5. Layout Rules

- **Single-column, mobile-first.** All primary screens (Dashboard, Add Contribution, Members, Expenses) are designed single-column first; wider viewports get more breathing room (max content width, centered) rather than a fundamentally different multi-column layout. There is no dedicated "desktop app" redesign — desktop is a wider version of the same layout, since the primary device is explicitly Sohail's phone (`docs/PRD.md` §Primary device).
- **Bottom navigation on mobile,** persistent and thumb-reachable, per `docs/APP_FLOW.md` §10. The single most important action — Add Contribution — gets the visually emphasized center slot.
- **Content max-width** on larger screens (~640–720px) so text lines and stat cards don't stretch uncomfortably wide on a desktop browser.
- **Spacing scale:** consistent 4px-based spacing scale (Tailwind defaults) — no ad hoc pixel values.
- **Sticky/fixed primary action** on long forms (e.g., Save button pinned to the bottom of the viewport on Add Contribution) so the admin never has to scroll to complete the core action — directly serves the "faster than the notebook" success metric.

## 6. Mobile-First Behavior (detail)

- **Touch targets:** minimum 44×44px for any tappable element (buttons, checklist rows, nav icons) — standard mobile accessibility minimum, doubly important for older users with less precise motor control.
- **Numeric keypad** (`inputmode="numeric"`) for all amount fields — no general keyboard forcing the user to hunt for number keys.
- **Quick-amount chips** are large, evenly spaced tap targets, not small text links.
- **No hover-dependent interactions.** Nothing in the app relies on a `:hover` state to reveal information or controls, since touch has no hover — this is a correctness rule, not a style preference.
- **Minimal typing.** Wherever a selection can replace typing (member search/select, quick-amount chips, month defaulting to current), it does — every keystroke saved serves the core speed metric.
- **One-handed reachability.** Primary actions live in the bottom half of the screen; anything requiring a stretch to the top of a large phone screen is secondary (e.g., logout, language switch in the header, not the footer).

## 7. Dashboard Structure (visual)

Reiterating the content order from `docs/APP_FLOW.md` §2 with visual intent:

1. Header — minimal, group name + month, not a hero banner.
2. Four stat cards in a 2×2 grid on mobile (single row on wider screens), each with a bilingual label, the number large and tabular, and a small icon (deposit/withdrawal/scale icons rather than generic dashboard iconography).
3. Add Contribution button — visually the most prominent interactive element on the page after the numbers themselves; solid fill, full width.
4. Paid/unpaid checklist — a simple list, not a data table; each row is a full-width tappable target (per touch-target rule) that deep-links into Add Contribution.
5. Recent activity — a lightweight list, de-emphasized (smaller text, muted color) relative to the stat cards and checklist, since it's supporting context, not the primary task.

## 8. Bilingual & RTL Handling

- **Labels:** core financial terms (Deposit/जमा, Expense/खर्च, Balance/बाकी) appear bilingually on stat cards and key headings regardless of the selected UI language, since these three words are the ones a quick glance most needs to recognize instantly for a mixed-literacy, mixed-familiarity user base. Full sentence-level content (buttons, form labels, help text) follows the selected language only, not bilingually duplicated everywhere — duplicating everything would clutter the calm layout this brief calls for.
- **Language switcher:** a small, always-visible icon/control in the header (globe icon + current language code), consistent placement across all screens per `docs/APP_FLOW.md` §10.
- **RTL support:** when Arabic is selected, the layout mirrors (nav, icons, text alignment) using CSS logical properties (`margin-inline-start` etc.) rather than hardcoded left/right values, so the same Tailwind-based components support LTR and RTL without a separate template set.
- **Script rendering:** typography stack (§3) includes RTL/Devanagari-capable fallback fonts so Arabic/Hindi render correctly without extra font-loading configuration per language.

## 9. UX Principles (summary)

1. **Speed of entry beats everything else** on the Add Contribution screen — this is the app's reason for existing (`docs/PRD.md` §8).
2. **Never show a number without context.** Every stat has a label; every list row shows who/when (audit trail is visible, not just logged).
3. **Confirm, don't celebrate.** Success states are quiet (a small toast), not attention-grabbing animation — appropriate tone for handling community trust money.
4. **Destructive actions always confirm**, non-destructive actions never make the user confirm unnecessarily (per `docs/APP_FLOW.md` §10) — avoid confirmation fatigue that trains users to tap "yes" without reading.
5. **Never hide the truth to look better.** If the group has had a bad month (spent more than collected), the dashboard shows that plainly (e.g., balance in `--color-danger` if negative) rather than obscuring it — transparency is a stated non-negotiable requirement (`docs/PRD.md` §3).

## 10. Accessibility

The rules in §2 (contrast), §3 (type size/weight), and §6 (touch targets, no hover-dependent interactions) already do most of the accessibility work required, because "readable for older users" and "accessible" overlap heavily for this audience. Full requirements, rationale, and a testing checklist live in `docs/ACCESSIBILITY.md` — target standard WCAG 2.1 AA. Two additions this brief didn't previously state explicitly, now required:

- **Visible focus indicators** on every focusable element (buttons, links, form fields, nav items), meeting the same 3:1 contrast minimum as icons — relevant for keyboard users and not just the touch-first mobile case this brief otherwise emphasizes.
- **Text must remain usable at 200% browser zoom** without clipping or overlapping content, since some older users rely on browser zoom rather than any in-app text-size control — no fixed-height containers that would clip enlarged text.

See `docs/ACCESSIBILITY.md` for the full commitment list (keyboard navigation, screen reader support, alt text on receipt photos, `prefers-reduced-motion`) and the verification checklist (Lighthouse, axe, keyboard-only pass, screen reader spot check).

## Assumptions

- Exact hex values in §2 are a starting palette or a v1 implementer/designer to adjust for measured contrast ratios (WCAG AA at minimum, given the older-user audience); the palette's *relationships* (calm, teal-forward, terracotta for expense, restrained accent use) are the fixed part of this brief, not the literal hex codes.
- Icon set assumed to be a single consistent icon library (e.g., Lucide, which pairs naturally with HTMX/templ/Tailwind stacks and has no framework lock-in) — exact library choice is an implementation detail, not a design decision.
- No dark mode is specified for v1; it can be added later as a CSS-variable theme swap without restructuring components, since colors are already tokenized (§2).
