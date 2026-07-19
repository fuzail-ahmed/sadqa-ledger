> **This is a template, not legal advice.** It is written to be filled in and published by the group running this software ("the operator"). It is not a substitute for advice from a qualified lawyer, especially if your group is large, operates across borders, or handles unusually sensitive information. If in doubt, consult a professional familiar with your local law.

# Privacy Policy — {{MASJID_NAME}}

*Last updated: {{DATE}}*

This page explains what information {{MASJID_NAME}} records when it uses the Sadqa Ledger software to track contributions and expenses, why it's recorded, who can see it, and what rights you have over your own information.

## Who is responsible for this data

{{MASJID_NAME}} runs its own copy of this software on its own server. **{{MASJID_NAME}} is the one responsible for this data — not the people who wrote the Sadqa Ledger software.** The software's authors never receive, store, or have access to any information entered into this instance; see the software's own [Disclaimer](https://github.com/fuzail-ahmed/sadqa-ledger/blob/main/legal/DISCLAIMER.md) if you'd like to read more about that distinction.

If you have a question or concern about your information, contact:

- **Contact person:** {{CONTACT_PERSON_NAME}}
- **Email:** {{CONTACT_EMAIL}}
- **Phone:** {{CONTACT_PHONE}}

## What we record

| Information | Recorded for | Who enters it |
|---|---|---|
| Your name | So your contributions can be tracked under your name, and so the group knows who has and hasn't paid this month | An admin, when you're added as a member |
| Phone number or household note *(optional)* | Only if the admins choose to note it, for their own reference (e.g., to contact you about payment) | An admin, optionally, as a free-text note |
| Contribution amount and date | To record what you gave and when, and to calculate the group's monthly and overall totals | An admin, at the time you pay |
| Expense records (description, amount, date, and an optional photo of the receipt) | To record what the group spent and on what, so spending is accountable | An admin |
| Admin account details (username, display name, a securely hashed password, and login session information such as approximate login time) | So the 2–3 people who manage {{MASJID_NAME}}'s ledger can log in and are each accountable for what they record | The admin themselves, when their account is created |

We do not record your religion, caste, income source, or any other information beyond what's listed above. We do not track your device, browsing habits, or location, beyond what a normal website already receives from your browser when you visit (like any website).

## Why we record it

To replace a paper notebook with an accurate, shared record of contributions and expenses — so the group's finances are correct, transparent, and not dependent on one person's memory or a notebook that could be lost. This is explained in more detail in the software's own [product documentation](https://github.com/fuzail-ahmed/sadqa-ledger/blob/main/docs/PRD.md), which anyone is welcome to read.

## Who can see what

- **Admins** ({{CONTACT_PERSON_NAME}} and the other 2–3 people who manage the ledger) can see everything: every member's name, contribution history, and all expense records.
- **The public transparency page** (a link anyone can be given, without needing to log in) shows the group's overall totals and recent activity. Whether your name is shown next to your contribution on this public page depends on a setting {{MASJID_NAME}} controls:
  - **If names are hidden (the default):** the public page shows totals only — how much came in, how much went out, the balance — with no individual names attached to amounts.
  - **If names are shown:** the public page lists contributions with the member's name next to the amount.
  - {{MASJID_NAME}} has currently set this to: **{{NAMES_PUBLIC_YES_OR_NO}}**. If you'd like this changed, or if you'd like to know before it changes, contact {{CONTACT_PERSON_NAME}} above.
- **Nobody outside {{MASJID_NAME}}** — not the software's authors, not any company — has access to this data. It lives only on the server {{MASJID_NAME}} runs, and in {{MASJID_NAME}}'s own backup storage (see below).

## Where this data lives, and where it doesn't go

This software is self-hosted: {{MASJID_NAME}} runs its own copy on its own server (or a server it has chosen and pays for, e.g., a cloud provider). Your information is stored in a single database file on that server. A copy is also continuously backed up to {{MASJID_NAME}}'s own private cloud storage account, so the data isn't lost if the server fails.

**No data is sent to the people who wrote this software, or to any third party for advertising, analytics, or resale.** The only place your data goes, besides the server itself, is {{MASJID_NAME}}'s own backup storage — a service {{MASJID_NAME}} pays for (or uses a free tier of) and controls the login credentials to.

## How long we keep it

Contribution and expense records are kept indefinitely as the group's permanent financial history, in the same way a paper ledger would traditionally be kept for years. If a member becomes inactive (e.g., moves away, stops contributing), their past records are kept for the group's accuracy and audit trail, but they're removed from the "who's paid this month" checklist and no longer treated as an active member.

## Your rights

You have the right to:

- **Ask what we have.** Contact {{CONTACT_PERSON_NAME}} to ask what information is recorded about you.
- **Ask us to correct it.** If your name is misspelled or an amount was logged wrong, tell an admin and it will be corrected — every correction is itself recorded (who fixed it and when), which is how this software keeps the ledger trustworthy.
- **Ask us to delete it.** You can ask to have your personal details (like a phone number note) removed. Note that removing your entire contribution history may affect the group's financial records and audit trail — {{CONTACT_PERSON_NAME}} will explain what can and can't be fully deleted before acting on your request, and will let you know within a reasonable time what was done.
- **Ask us to stop showing your name publicly**, even if the group's general setting is to show names — raise this with {{CONTACT_PERSON_NAME}}.

## Security measures we actually use

This isn't a list of promises — it's what the software this instance runs actually does:

- The website is only reachable over an encrypted connection (HTTPS), so data can't be read in transit.
- Passwords for the 2–3 admin accounts are never stored in readable form — they're stored using a one-way scrambling method (bcrypt) that can't be reversed even if the database were somehow read by someone unauthorized.
- Login sessions are similarly protected — the value that proves you're logged in is never stored in a form that could be reused if the database leaked.
- Every entry (contribution or expense) records which admin recorded it and when, so mistakes and changes are traceable.
- The database is backed up continuously to private, password-protected cloud storage that {{MASJID_NAME}} controls.

## Legal framework

{{MASJID_NAME}} is based in {{COUNTRY}}. Where {{MASJID_NAME}} operates in India, this policy is written with India's **Digital Personal Data Protection Act, 2023 (DPDP Act)** in mind as the relevant framework — in particular its expectations around collecting only necessary information, being clear about why it's collected, and letting people ask to access or correct their own data, all of which this policy tries to reflect in plain language above.

**If {{MASJID_NAME}} operates outside India, or has members outside India, please have this policy checked against your own country's data protection law** (for example, but not limited to, the GDPR in the EU/UK) — this template does not attempt to cover every country's requirements.

## Changes to this policy

If {{MASJID_NAME}} changes what it records or how it's used, this page will be updated and the date at the top will change. For significant changes (for example, turning on public name display), {{MASJID_NAME}} will make a reasonable effort to let members know directly, not just update this page silently.

---

## हिन्दी में संक्षिप्त सारांश

{{MASJID_NAME}} इस ऐप में आपका नाम, कभी-कभी फ़ोन नंबर (अगर एडमिन ने नोट किया हो), आपके चंदे की राशि और तारीख़, और खर्च का ब्यौरा दर्ज करता है। यह जानकारी सिर्फ़ 2-3 भरोसेमंद एडमिन देख सकते हैं। सार्वजनिक लिंक पर कुल राशि दिखती है, लेकिन नाम तभी दिखते हैं जब समूह ने यह सेटिंग जान-बूझकर चालू की हो (डिफ़ॉल्ट रूप से नाम छुपे रहते हैं)। यह डेटा सिर्फ़ {{MASJID_NAME}} के अपने सर्वर और उसकी अपनी बैकअप स्टोरेज में रहता है — यह ऐप बनाने वालों के पास कभी नहीं जाता। अगर आप अपनी जानकारी देखना, ठीक करवाना, या हटवाना चाहते हैं, तो {{CONTACT_PERSON_NAME}} ({{CONTACT_EMAIL}} / {{CONTACT_PHONE}}) से संपर्क करें।

## Assumptions

- Placeholders used: `{{MASJID_NAME}}`, `{{DATE}}`, `{{CONTACT_PERSON_NAME}}`, `{{CONTACT_EMAIL}}`, `{{CONTACT_PHONE}}`, `{{NAMES_PUBLIC_YES_OR_NO}}`, `{{COUNTRY}}` — an operator should be able to fill these in without technical help.
- The data inventory table is drawn directly from `docs/SCHEMA.md`'s `members`, `contributions`, `expenses`, and `admins` tables (§3.1–3.4); the `sessions` table is described in plain language as "login session information" rather than listing its technical columns, since members don't need column-level detail, only that logging in creates a short-lived, securely-handled record.
- `group_settings.show_names_publicly` is surfaced as the `{{NAMES_PUBLIC_YES_OR_NO}}` placeholder so the operator states their actual current setting rather than a generic description — this keeps the policy honest and specific rather than boilerplate.
- DPDP Act 2023 is treated as the default reference given the project's India origin (`docs/PRD.md`), with an explicit, prominent caveat for operators elsewhere, per the task's instructions — this document does not attempt a full GDPR-equivalent legal analysis, which would need a qualified lawyer for any group operating at meaningful scale in the EU/UK.
- The Hindi summary is a condensed, plain-language version, not a full translation of every section — chosen deliberately so it's actually readable by a village committee member rather than a literal machine-style translation of the full legal document.
