# Self-Hosting Guide — Deploying Your Own Sadqa Ledger

This guide assumes you have **never used a terminal seriously before**. Every step is spelled out. If a step doesn't work exactly as described, stop and re-read the previous step before continuing — most problems come from a skipped step, not a bug.

By the end, you'll have:

- Your own copy of Sadqa Ledger running on a **free** server, reachable at your own web address, with a padlock (HTTPS) in the browser.
- Automatic, continuous backups of your data to free cloud storage, so you never lose it even if the server itself is destroyed.

Total ongoing cost: **$0/month**, using free tiers designed to stay free indefinitely (not a trial).

Related reading if you want the "why" behind these choices: [`docs/TRD.md`](TRD.md) §Deployment and §Security.

---

## Overview of what you're building

```
Your phone/browser  →  HTTPS  →  Caddy (free automatic certificate)  →  Sadqa Ledger app  →  SQLite file
                                                                                                    ↓
                                                                                       Litestream (continuous backup)
                                                                                                    ↓
                                                                                       Cloudflare R2 (free storage)
```

You'll set up, in order:

1. A free Oracle Cloud server (a "VM" — virtual machine, a computer that lives in the cloud).
2. A domain name pointing at that server — a free DuckDNS subdomain is the recommended default (see Step 2's Domain options for the alternatives).
3. Docker on that server (the tool that runs the app).
4. Cloudflare R2 (free storage for backups).
5. The app itself, plus Caddy (handles HTTPS automatically) and Litestream (handles backups automatically).

---

## Step 1 — Create a free Oracle Cloud account and VM

1. Go to [oracle.com/cloud/free](https://www.oracle.com/cloud/free/) and sign up for an "Always Free" account. You'll need to enter a credit card for identity verification, but the resources in this guide stay within the permanently-free tier — Oracle will not charge you as long as you don't upgrade or exceed the free limits.
2. Once your account is active, go to the **OCI Console** (Oracle Cloud Infrastructure Console).
3. Navigate to **Compute → Instances → Create Instance**.
4. Name it something like `sadqa-ledger-server`.
5. Under **Image and shape**:
   - Choose an **Ampere (ARM) Always Free-eligible shape** (e.g., `VM.Standard.A1.Flex`) with the free allotment (up to 4 OCPUs / 24 GB RAM total across your Always Free VMs — this app needs far less; 1 OCPU / 6 GB is plenty).
   - Choose a recent **Ubuntu** image (e.g., Ubuntu 22.04 or newer) as the operating system.
6. Under **Networking**, keep the default VCN (virtual network) settings — Oracle will create one for you if you don't have one yet. Make sure **"Assign a public IPv4 address"** is checked.
7. Under **Add SSH keys**, choose **"Generate a key pair for me"** and click **Download private key**. Save this file somewhere safe on your computer (e.g., `sadqa-ledger-key.key`) — you'll need it to connect to your server. Do not lose this file and do not share it.
8. Click **Create**. Wait a few minutes for the instance to show status **Running**.
9. Note the **Public IP address** shown on the instance's details page — you'll use this to connect.

### Open the necessary ports

By default, Oracle's firewall blocks web traffic. You need to allow HTTP (port 80) and HTTPS (port 443) so Caddy can serve your site and get a certificate.

1. On your instance's details page, click the link under **Virtual cloud network** → then **Security Lists** (or **Network Security Groups**, depending on your setup) → your default security list.
2. Click **Add Ingress Rules** and add two rules:
   - Source CIDR: `0.0.0.0/0`, IP Protocol: TCP, Destination Port: `80`
   - Source CIDR: `0.0.0.0/0`, IP Protocol: TCP, Destination Port: `443`
3. Save.

### Connect to your server

On Windows, use **PowerShell**; on Mac/Linux, use **Terminal**.

```bash
chmod 600 sadqa-ledger-key.key
ssh -i sadqa-ledger-key.key ubuntu@<your-server-public-ip>
```

Replace `<your-server-public-ip>` with the IP address you noted above. Type `yes` if asked whether to trust the connection. You're now "inside" your cloud server.

---

## Step 2 — Point a domain name at your server

Caddy (Step 5) needs *some* address to serve your site at and, for a real HTTPS padlock, a domain name pointing at your server's IP. You have three options. Read the trade-offs below, pick one, and skip to the matching instructions.

### Domain options

| Option | Cost | Looks like | Best for |
|---|---|---|---|
| **1. Free DuckDNS subdomain** (recommended) | $0/year | `yourmasjid.duckdns.org` | Almost everyone — this is the default path this guide assumes from here on |
| **2. Cheap custom domain via Cloudflare** | ~$2–15/year (varies by domain ending) | `ledger.yourmasjid.org` | Groups that want a more official-looking address and don't mind a small yearly cost |
| **3. IP address only, no domain** | $0 | `http://140.x.x.x:8080` | Quick local testing only — **not recommended for real use** |

**Option 1 — Free DuckDNS subdomain (recommended $0 path).** [DuckDNS](https://www.duckdns.org) gives you a free subdomain (`yourmasjid.duckdns.org`) that points at your server's IP address. It works with Caddy's automatic HTTPS exactly the same way a paid domain does — Caddy doesn't care that it's free, it just needs a real domain name it can request a certificate for. The only downside is the address looks a little less "official" than your own `.org`, and it depends on DuckDNS staying online (it's a well-established free service, but it is a third party). This is the right choice for almost every group reading this guide, and it's what the rest of this document assumes unless you say otherwise below.

*Setup:*
1. Go to [duckdns.org](https://www.duckdns.org) and sign in (with Google, GitHub, or another supported account — no new password to remember).
2. Under "add domain", type a subdomain name, e.g. `yourmasjid` (this becomes `yourmasjid.duckdns.org`), and click **add domain**.
3. In the **current ip** field next to your new subdomain, enter your server's public IP address from Step 1, then click **update ip**.
4. That's it — `yourmasjid.duckdns.org` now points at your server. If your server's IP ever changes (it normally won't, since Oracle assigns a fixed IP to your instance), come back here and update it.
5. Use `yourmasjid.duckdns.org` wherever this guide says "your domain" below (for example, `BASE_URL=https://yourmasjid.duckdns.org` in Step 5).

**Option 2 — Cheap custom domain with Cloudflare DNS.** If your group wants an address that looks like `ledger.yourmasjid.org` rather than a `.duckdns.org` address, and is fine with a small yearly cost (often $2–15/year depending on the domain ending you pick, e.g. `.org`, `.in`, `.com` — this is separate from the $0 hosting target and paid to a domain registrar, not to us or to Oracle), this is the option for you. You'll register the domain, then use Cloudflare's free plan to manage where it points.

*Setup:*
1. Register a domain through any registrar — [Cloudflare Registrar](https://www.cloudflare.com/products/registrar/) sells at-cost with no markup, or use one you already have (Namecheap, GoDaddy, etc.).
2. Sign up for a free account at [dash.cloudflare.com](https://dash.cloudflare.com) and add your domain there (Cloudflare will give you two nameservers to set at your registrar — this points your domain's DNS management to Cloudflare, it doesn't move your registration).
3. In Cloudflare's DNS settings for your domain, add an **A record**:
   - Name: `ledger` (or whatever subdomain you want, e.g. `ledger.yourmasjid.org`)
   - IPv4 address: your server's public IP from Step 1
   - **Proxy status: set this to "DNS only" (grey cloud, not the orange "Proxied" cloud).** This is important — if Cloudflare's orange-cloud proxy is on, it intercepts traffic before it reaches Caddy and Caddy won't be able to get its own certificate. Grey cloud means Cloudflare only handles DNS, and your traffic goes straight to your server, exactly like Option 1.
4. Wait for DNS to propagate (usually a few minutes, occasionally up to a few hours). Check with [whatsmydns.net](https://www.whatsmydns.net/).
5. Use `ledger.yourmasjid.org` (your actual domain) wherever this guide says "your domain" below.
6. **Set a calendar reminder before your domain's renewal date** — unlike Option 1, a custom domain lapses if not renewed, and losing it means your public link stops working until you re-register and re-point it.

**Option 3 — IP address only (testing only, not recommended for real use).** You can skip domain setup entirely and reach your server directly at `http://<your-server-ip>:8080`. This is useful only for a quick local test before you commit to a domain. It has real downsides for actual use:
- **No HTTPS padlock is possible this way** — Caddy's automatic certificate feature requires a real domain name; an IP address can't get one. Your login password would travel unencrypted, which is not acceptable for a real deployment handling real community data.
- The address is harder to remember and share than a domain name.
- If your server's IP ever changes, the address changes too, with nothing to update automatically (a domain just gets re-pointed; an IP-only setup breaks).

If you're only kicking the tires before deciding, this is fine — skip straight to Step 3 and come back to Step 2 once you're ready to go live. **Before recording any real member's data, upgrade to Option 1 or 2.**

---

## Step 3 — Install Docker on your server

Still connected via SSH from Step 1, run:

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER
```

Log out and back in (close the SSH window and reconnect with the same `ssh` command from Step 1) so the Docker permission takes effect. Confirm it worked:

```bash
docker --version
```

You should see a version number, not an error.

---

## Step 4 — Set up free backup storage (Cloudflare R2)

1. Sign up at [dash.cloudflare.com](https://dash.cloudflare.com) (free account).
2. In the sidebar, go to **R2 Object Storage** → **Create bucket**. Name it something like `sadqa-ledger-backups`. Leave other settings default.
3. Go to **Manage R2 API Tokens** (in R2 settings) → **Create API Token**. Give it **Object Read & Write** permission, scoped to your new bucket if possible. Save the **Access Key ID** and **Secret Access Key** shown — you won't be able to see the secret again after leaving this page.
4. Note your **Account ID** (shown in the R2 dashboard URL or account settings) — you'll need it to build the R2 endpoint URL: `https://<account-id>.r2.cloudflarestorage.com`.

---

## Step 5 — Configure and run Sadqa Ledger

Still on your server (via SSH):

```bash
mkdir -p ~/sadqa-ledger && cd ~/sadqa-ledger
curl -o .env.example https://raw.githubusercontent.com/fuzail-ahmed/sadqa-ledger/main/.env.example
cp .env.example .env
nano .env
```

`nano` is a simple text editor. Fill in every value in `.env` — see the comments in the file for what each one means. At minimum you must set:

- `BASE_URL` — `https://yourmasjid.duckdns.org` (or your custom domain, whichever option you chose in Step 2 — use `http://<your-server-ip>:8080` only if you're on Option 3 for quick testing, with no HTTPS)
- `SESSION_SECRET` — generate one with `openssl rand -hex 32` (run that command, copy the output in)
- `LITESTREAM_R2_BUCKET`, `LITESTREAM_R2_ENDPOINT`, `LITESTREAM_R2_ACCESS_KEY_ID`, `LITESTREAM_R2_SECRET_ACCESS_KEY` — from Step 4

Save and exit `nano` with `Ctrl+O`, `Enter`, then `Ctrl+X`.

Now download the deployment files and start everything:

```bash
curl -o docker-compose.yml https://raw.githubusercontent.com/fuzail-ahmed/sadqa-ledger/main/docker-compose.yml
curl -o Caddyfile https://raw.githubusercontent.com/fuzail-ahmed/sadqa-ledger/main/Caddyfile
docker compose up -d
```

This starts three things together: the Sadqa Ledger app, Litestream (continuous backup), and Caddy (HTTPS). If you chose Option 1 or 2 in Step 2, Caddy will automatically request a free HTTPS certificate for your domain the first time it starts — this requires your DNS (Step 2) to already be pointing at the server, and ports 80/443 to be open (Step 1). If you're on Option 3 (IP-only testing), Caddy has no domain to get a certificate for, so you'll reach the app directly over plain HTTP instead — expect a browser warning, and don't use this for real data (see Step 2, Option 3).

Check everything is running:

```bash
docker compose ps
```

You should see all services listed as `Up`/`running`.

Visit your address from Step 2 in a browser — `https://yourmasjid.duckdns.org`, your custom domain, or `http://<your-server-ip>:8080` for Option 3. You should see the Sadqa Ledger first-run setup screen. Follow it to create your first admin account and set your group's name and currency.

---

## Step 6 — Verify backups are working

Give it a few minutes after first startup, then check the Litestream logs:

```bash
docker compose logs litestream
```

You're looking for lines indicating a successful replication to your R2 bucket, without repeated error messages. You can also check your Cloudflare R2 bucket in the dashboard — you should see files appearing under a path matching your database name.

**It's worth doing a real restore test once, now, while there's little data to lose** — see the next section — so you know the process works before your group depends on it.

---

## Step 7 — Before you invite your community

Your server is running, but before you start recording real members' names and contributions, take a few minutes to do this — it matters more than any of the technical steps above:

1. **Fill in your privacy policy.** Copy [`legal/PRIVACY_POLICY_TEMPLATE.md`](../legal/PRIVACY_POLICY_TEMPLATE.md), fill in your masjid's name, contact details, and current privacy setting, and publish it somewhere your members can read it (a printed notice, a WhatsApp message, a page on the public site — whatever fits your community).
2. **Do the same with [`legal/TERMS_TEMPLATE.md`](../legal/TERMS_TEMPLATE.md).**
3. **Read [`docs/OPERATOR_RESPONSIBILITIES.md`](OPERATOR_RESPONSIBILITIES.md) in full.** It's short, plain-language, and covers the things that actually matter day to day — strong admin passwords, deciding your name-privacy setting deliberately, knowing what to do if a member asks about their data, and applying updates when they're released.

None of this requires a lawyer for a typical single-masjid instance, but it does require you, the operator, to actually read these two short documents once before going live — this is the part of "self-hosting" that's on you, not the software.

---

## Restore procedure (disaster recovery)

If your server is lost, corrupted, or you're moving to a new server, restore from your R2 backup like this:

1. Set up a new server (or reuse the existing one after stopping the app: `docker compose down`).
2. Make sure your `.env` file has the correct `LITESTREAM_R2_*` values pointing at the same bucket that was being backed up.
3. Run Litestream's restore command before starting the app:

```bash
docker run --rm \
  --env-file .env \
  -v sadqa-ledger-data:/data \
  litestream/litestream restore -o /data/sadqa-ledger.db \
  "s3://${LITESTREAM_R2_BUCKET}/sadqa-ledger.db"
```

(Litestream reads the R2 credentials and endpoint from the same environment variables already in your `.env` file.)

4. Once the restore command finishes successfully, start the full stack normally:

```bash
docker compose up -d
```

5. Log in and confirm your members, contributions, and expenses are all present and match what you expect.

**Important:** the database and CSV files an admin downloads from the app's own `/export` screen (see [`docs/APP_FLOW.md`](APP_FLOW.md) §8) are a *different*, sanitized copy — they deliberately exclude login sessions and password hashes, and are meant for manual backup/migration by an admin, not as the primary disaster-recovery mechanism. Litestream/R2 (this section) is the automatic, complete backup that should always be running in the background.

---

## Upgrading to a new version

When a new version of Sadqa Ledger is released:

```bash
cd ~/sadqa-ledger
docker compose pull
docker compose up -d
```

This pulls the newest published image and restarts the app with it. The app applies any needed database migrations automatically on startup (see [`docs/SCHEMA.md`](SCHEMA.md) §8) — you do not need to run anything manually. Check the release notes/[`CHANGELOG.md`](../CHANGELOG.md) for anything unusual before upgrading, especially for major version changes.

To roll back if something goes wrong, edit `docker-compose.yml` to pin the previous version tag instead of `latest`, then run `docker compose up -d` again.

---

## Troubleshooting

- **Browser shows "connection refused" or times out:** confirm DNS has propagated (Step 2) and that ports 80/443 are open in Oracle's security list (Step 1). Run `docker compose ps` to confirm containers are running.
- **Caddy can't get a certificate:** this almost always means DNS isn't pointing at your server yet, ports 80/443 aren't open, or (Option 2) Cloudflare's proxy is still set to "Proxied" (orange cloud) instead of "DNS only" (grey cloud). Check `docker compose logs caddy` for the specific error.
- **"Permission denied" running `docker` commands:** you likely need to log out and back in after Step 3's `usermod` command, or prefix commands with `sudo` as a temporary workaround.
- **Litestream errors mentioning credentials:** double-check your R2 Access Key ID/Secret in `.env` — they're easy to mistype when copy-pasting.
- Still stuck? Open an issue using the bug report template in the repository — include what step you're on and the exact error text (with any secrets blanked out).
