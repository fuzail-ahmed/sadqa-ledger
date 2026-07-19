> This is a practical checklist, not legal advice. For anything high-stakes — a data request you're unsure how to handle, a serious security incident, a legal question about your group's obligations — talk to a qualified professional. This page just tells you, in plain language, what running this software makes you responsible for day to day.

# Your Responsibilities as an Operator

If you're the person who set up Sadqa Ledger for your masjid or group, this page is for you. It's written for a volunteer, not a technical expert or a lawyer — no jargon, just what you need to know and do.

## The big picture

**This software runs on a server you (or your group) control. That means the data in it — your members' names, contributions, expenses — is your group's responsibility, not the software's authors'.** The people who wrote Sadqa Ledger never see this data and can't help you manage it day to day. You're the one who can. This isn't meant to feel heavy — it's the same responsibility Sohail already had with the paper notebook, just with a few new things to set up once and then maintain occasionally.

Use this page as a running checklist, not a one-time read.

## Before you start recording real data

- [ ] **Tell your members what you're recording, and why.** Even a short spoken announcement at a gathering ("we're moving to a simple app to track contributions, here's what it records") builds trust. Better yet, fill in and share `legal/PRIVACY_POLICY_TEMPLATE.md` and `legal/TERMS_TEMPLATE.md` with your group.
- [ ] **Decide your name-privacy setting before you turn it on for real.** The public page can either hide member names (default) or show them next to amounts. Once you show names publicly, anyone with the link has seen them — you can turn the setting back off, but you can't un-show what was already visible. Think about it once, deliberately, rather than toggling it to "see what it looks like."
- [ ] **Use strong, unique admin passwords.** Each of your 2–3 admins should have their own password that isn't reused from another account and isn't something easy to guess (a birthday, "masjid123", etc.). A password manager helps, but even writing a long random phrase down somewhere safe is fine.
- [ ] **Make sure HTTPS is working** (the padlock icon in the browser, and the site address starting with `https://`). This is set up automatically if you followed `docs/DEPLOY.md`'s Caddy step — just confirm it's actually showing the padlock before you rely on the site for real data.
- [ ] **Set up backups and actually test them once.** `docs/DEPLOY.md`'s Litestream section walks through this. Don't skip the restore test — a backup you've never restored from is a backup you don't actually know works.

## Ongoing responsibilities

- [ ] **Only share the public link as widely as you intend.** It doesn't require a login, so anyone who has the link can view it (and, if you've set names to public, can see them). Share it the way you'd share something semi-private — directly with your community, not posted somewhere fully public like an open social media post, unless that's genuinely your intent.
- [ ] **Keep your admin list current.** If an admin stops helping with the ledger, remove their account (Settings → Admins) rather than leaving it active indefinitely.
- [ ] **Apply software updates when they're released.** See `SECURITY.md` — nobody but you can update your running instance. Check for new releases occasionally, especially if one is flagged as fixing a security issue.
- [ ] **Correct mistakes openly rather than hiding them.** Every entry and edit is recorded with who made it — that's a feature, not something to work around. If a number's wrong, fix it and let the audit trail show the correction.

## If a member asks about their data

You don't need to be a lawyer to handle this well — just be straightforward:

- [ ] **If they ask what you have on them:** tell them. It's just their name, maybe a phone number note, and their contribution history — you can look at their member record together if that's easiest.
- [ ] **If they ask you to fix something:** fix it. Correcting a name or amount is a normal admin action.
- [ ] **If they ask you to delete something:** you can remove personal notes (like a phone number) directly. For their contribution history, explain that removing it may affect the group's financial record and audit trail, and agree together on what makes sense — full removal, or marking them inactive so future entries stop but history is kept for the group's own accuracy. There's no single right answer here; use judgment, and see `legal/PRIVACY_POLICY_TEMPLATE.md` for how to phrase your group's own policy on this.
- [ ] **If they ask you to stop showing their name publicly**, even though your group's general setting shows names, honor that for them individually going forward.

## Understanding the name-privacy setting (worth re-reading)

This is the one setting most likely to have real consequences if flipped casually:

- **Off (default):** the public page and WhatsApp summaries show totals only — never who gave what.
- **On:** anyone with the public link sees exactly who contributed and how much, for every entry.

Some people give quietly on purpose — that's a value worth respecting even if most of your group is comfortable with names being shown. If you're not sure, leave it off and ask your community directly before turning it on.

## A quick reality check

None of this is meant to make you feel like you need a legal team to run a masjid ledger app. Most of it is common sense you'd apply to any notebook of community information — this page just makes it explicit because a website reaches further and faster than a notebook in a drawer. If something on this list feels like too much, that's worth a conversation with the rest of your admins, not something to just skip silently.

## Related reading

- `legal/PRIVACY_POLICY_TEMPLATE.md` — fill this in and publish it for your members.
- `legal/TERMS_TEMPLATE.md` — short terms to share alongside the privacy policy.
- `legal/DISCLAIMER.md` — what the software's authors are and aren't responsible for.
- `docs/DEPLOY.md` — the technical setup steps referenced above (HTTPS, backups).
- `SECURITY.md` — how to report a vulnerability, and why updates are your responsibility.

## Assumptions

- This checklist assumes the reader has already completed the technical setup in `docs/DEPLOY.md` and is now operating the instance day to day — it deliberately doesn't repeat DEPLOY.md's step-by-step server setup instructions, only reminds the operator of the responsibilities that setup created.
- "Deleting a member's contribution history" is presented as a judgment call between full removal and marking inactive, rather than a strict rule, because the right answer genuinely depends on the group's own audit/accuracy needs versus the individual's request — a rigid rule here risked being wrong in either direction.
