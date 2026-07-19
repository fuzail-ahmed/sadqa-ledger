> This document is not legal advice, for either operators or the software's authors. It states our position as the software's authors plainly; if you need a formal legal opinion on any of this, consult a qualified professional.

# Disclaimer — Sadqa Ledger Software Authors

This disclaimer is from the people who write and maintain the Sadqa Ledger open-source project. It's separate from — and should not be confused with — the privacy policy or terms that the operator of a specific instance (e.g., a masjid or community group) publishes for their own members using [`legal/PRIVACY_POLICY_TEMPLATE.md`](PRIVACY_POLICY_TEMPLATE.md) and [`legal/TERMS_TEMPLATE.md`](TERMS_TEMPLATE.md).

## We are software authors, not a service provider

Sadqa Ledger is self-hosted, open-source software distributed under the MIT License. We write the code and publish it; **we do not run any instance of it, host any group's data, or receive any information entered into any instance.** Every masjid or group that uses this software downloads it, runs it on infrastructure of their own choosing, and operates it themselves.

This means:

- **We are not a data controller or data processor** for any operator's members. We have no technical access to any running instance's database, backups, or logs, and no mechanism to request it.
- **We cannot see, export, or delete anyone's data.** If you're a member of a group using this software and want your information corrected or removed, the people to contact are that group's admins — see `docs/OPERATOR_RESPONSIBILITIES.md` for what we ask operators to be ready to do when a member asks.
- **We are not party to any agreement between an operator and its members.** The operator's own privacy policy and terms (drafted from our templates or otherwise) govern that relationship, not us.

## No warranty

Consistent with the project's MIT License:

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

In plain language: we make this available in good faith, and we try to build it carefully — see `docs/TRD.md` for the security and accuracy reasoning behind the design — but we cannot promise it is error-free, uninterrupted, or fit for every possible use, and we are not liable for losses arising from its use.

## This is not accounting or financial software

Sadqa Ledger is a simple contribution-and-expense ledger, built to replace a paper notebook with something more accurate and durable — see `docs/PRD.md` for its stated scope. It is **not**:

- Certified accounting software.
- A replacement for formal bookkeeping, statutory audit, or financial reporting where your local law, trust deed, society registration, or tax status requires one.
- Tax, legal, or financial advice of any kind.

If your group is legally required to maintain audited accounts, file returns, or meet specific financial record-keeping standards (as many registered trusts, societies, or religious organizations are, depending on jurisdiction), **this software can be a useful day-to-day record but does not by itself satisfy those obligations.** Please check with an accountant or the relevant authority for your organization's specific legal form and jurisdiction.

## Operators are solely responsible for their own compliance and data

Each operator (the masjid or group running its own instance) is solely responsible for:

- Complying with data protection, privacy, and any other applicable law in their own jurisdiction (see `legal/PRIVACY_POLICY_TEMPLATE.md`'s note on the DPDP Act 2023 and other frameworks).
- Securing their own server, backups, and admin accounts (see `docs/DEPLOY.md` and `docs/OPERATOR_RESPONSIBILITIES.md`).
- Deciding what to record, who can see it, and how long to keep it.
- Responding to their own members' requests about their data.
- Their own group's financial record-keeping and reporting obligations, if any.

We provide the software and documentation to make good decisions on all of the above easier — we cannot make those decisions or guarantee compliance on any operator's behalf.

## Questions

If you're an operator and something in this disclaimer, the privacy policy template, or the terms template is unclear, opening an issue on the repository is welcome for anything about how the *software* works. For anything about your specific group's legal obligations, please speak to a qualified professional in your jurisdiction instead — we're not able to give that kind of advice.

## Assumptions

- This disclaimer intentionally does not attempt to restate every clause of the MIT License in full — it quotes the operative no-warranty/no-liability paragraph and points to `LICENSE` for the complete text, to avoid two documents drifting out of sync over time.
- "Accounting or financial software" boundary is described in general terms (audited accounts, statutory filings) rather than naming specific jurisdictions' requirements (e.g., a specific Indian Trust Act or Wakf Board rule), since those vary by country, state, and organizational registration type — an operator needing that specificity is directed to a qualified accountant rather than given potentially wrong specifics here.
