# Self-Hosting Guide - Deploying Sadqa Ledger on Oracle Cloud

This guide deploys Sadqa Ledger on an Oracle Cloud Ubuntu VM using Docker Compose. By default, it pulls prebuilt Docker images from GitHub Container Registry (GHCR) to avoid building on the server.

By the end, you will have:

- Caddy serving the app over HTTPS with an automatic certificate.
- Sadqa Ledger running from a prebuilt GHCR Docker image.
- SQLite stored on a persistent Docker volume.
- Litestream continuously replicating the SQLite database to Cloudflare R2.

## Architecture

```text
Browser -> HTTPS -> Caddy -> Sadqa Ledger (GHCR Image) -> /data/sadqa-ledger.db
                                                               |
                                                               v
                                                          Litestream -> Cloudflare R2
```

The Compose stack is defined in `docker-compose.yml` and starts three services:

- `app`: prebuilt image pulled from GitHub Container Registry (GHCR).
- `caddy`: public HTTP/HTTPS entrypoint.
- `litestream`: SQLite backup replication sidecar.

## 1. Create the Oracle Cloud VM

Create an Oracle Cloud Always Free Ubuntu VM. The ARM Ampere shape works well and is the recommended free option.

Open these ingress ports in the instance security list or network security group:

- TCP `80`
- TCP `443`

Then SSH into the VM:

```bash
ssh -i sadqa-ledger-key.key ubuntu@<your-server-public-ip>
```

## 2. Point DNS at the VM

Create an `A` record for your domain or subdomain pointing to the VM public IP.

Example:

```text
madni-app.cocoontrix.com -> <your-server-public-ip>
```

If using Cloudflare DNS, keep the record as **DNS only** while Caddy obtains and renews certificates. Do not enable the orange-cloud proxy for the first deployment.

## 3. Install Docker and Git

On the VM:

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker "$USER"
```

Log out and back in so the Docker group membership applies, then verify:

```bash
docker --version
docker compose version
```

## 4. Clone the repository

```bash
git clone https://github.com/fuzail-ahmed/sadqa-ledger.git
cd sadqa-ledger
```

For a private fork, use your fork URL instead.

## 5. Configure environment

```bash
cp .env.example .env
nano .env
```

Set at least these values:

```env
DOMAIN=madni-app.cocoontrix.com
BASE_URL=https://madni-app.cocoontrix.com
SESSION_SECRET=<output of: openssl rand -hex 32>
DATABASE_PATH=/data/sadqa-ledger.db
```

`DOMAIN` is the bare hostname Caddy uses as its site label. Do not include `https://`.

`BASE_URL` is the full public URL the app uses for absolute links and secure-cookie behavior. It should include `https://`.

Set the Litestream R2 variables if you want production backups:

```env
LITESTREAM_R2_BUCKET=your-r2-bucket-name
LITESTREAM_R2_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
LITESTREAM_R2_ACCESS_KEY_ID=your-r2-access-key-id
LITESTREAM_R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
```

Generate the session secret on the VM:

```bash
openssl rand -hex 32
```

## GitHub Container Registry (Default & Recommended)

By default, the `docker-compose.yml` file is configured to pull prebuilt Docker images of Sadqa Ledger from GitHub Container Registry (GHCR) (`ghcr.io/fuzail-ahmed/sadqa-ledger:latest`). These images support both `linux/amd64` and `linux/arm64` architectures.

### Why GHCR is the Default
- **No Heavy Compilations on VM:** Compiling Go code, generating templ files, and minifying Tailwind CSS is handled in GitHub Actions instead of on the server.
- **Fast Deployments:** Launching the application takes seconds since you only pull prebuilt layers instead of building from source.
- **Multitarget Architecture Support:** Seamlessly runs on Oracle Cloud ARM64 Ampere instances and traditional AMD64 hosts.

### Why Building on 1 GB Oracle VMs is Discouraged
Oracle Cloud's Always Free tier provides VM instances (like `VM.Standard.E2.1.Micro` or low-end Ampere shapes) with limited memory (e.g., 1 GB RAM). Running `docker compose build` on these VMs is highly discouraged because:
- **Out of Memory (OOM) Errors:** The Go compiler, Tailwind CLI, and templ generators require significant RAM. Running them on a 1 GB VM frequently exhausts memory, causing the OS to kill the compilation process.
- **System Instabilities:** The high CPU and memory pressure during builds can cause the VM to freeze, terminating existing services like Caddy or Litestream backups.

### How to Roll Back to a Previous Image Tag
If a new release has issues, you can easily roll back to a specific tag or git SHA without editing the `docker-compose.yml` file:
1. Open `.env` and configure `APP_VERSION` to point to the desired release tag (e.g., `v1.0.3`) or git SHA (e.g., `3d9e81f`):
   ```env
   APP_VERSION=v1.0.3
   ```
2. Pull the configured image version and restart the containers:
   ```bash
   docker compose pull
   docker compose up -d
   ```

### Local Development Builds
If you are developing locally and want to build the Docker image from your local source files instead of pulling from GHCR:
1. Copy the example override configuration:
   ```bash
   cp docker-compose.override.yml.example docker-compose.override.yml
   ```
2. When running Docker Compose commands locally, Docker Compose will automatically read both `docker-compose.yml` and `docker-compose.override.yml`, merging them to build the application from source.

---

## 6. Pull and start

To start the deployment:

```bash
docker compose pull
docker compose up -d
```

> [!NOTE]
> Production deployments pull prebuilt images from GitHub Container Registry (GHCR) by default. This avoids compiling code on the server, which is essential for low-memory VMs (e.g. 1 GB RAM) to prevent memory exhaustion and system instability.



Check status:

```bash
docker compose ps
docker compose logs --tail=100 app
docker compose logs --tail=100 caddy
docker compose logs --tail=100 litestream
```

Open your site:

```text
https://madni-app.cocoontrix.com
```

You should see the first-run setup screen. Create the first admin account and configure the group.

## 7. Verify backups

After the app has started and the first database file exists, check Litestream:

```bash
docker compose logs litestream
```

You should see replication activity without repeated credential or bucket errors.

Litestream reads the same SQLite file as the app through the shared `sadqa-ledger-data` volume. The app opens SQLite in WAL mode, and Litestream is designed to replicate SQLite WAL changes from the filesystem.

## Restore Procedure

Stop the stack before restoring:

```bash
docker compose down
```

Load `.env` into the current shell so the restore URL expands, then restore into the same named Docker volume:

```bash
set -a
. ./.env
set +a

docker run --rm \
  --env-file .env \
  -v sadqa-ledger-data:/data \
  litestream/litestream:0.3 restore \
  -o /data/sadqa-ledger.db \
  "s3://${LITESTREAM_R2_BUCKET}/sadqa-ledger.db"
```

Then restart:

```bash
docker compose up -d
```

Log in and confirm the expected members, contributions, expenses, settings, and public page are present.

## Production Update

To update the application to the newest version on production:

```bash
git pull
docker compose pull
docker compose up -d
docker compose ps
```

This pulls the prebuilt image configured via `APP_VERSION` in your `.env` file (which defaults to `latest`) and restarts the stack with minimal downtime. The application automatically runs any pending migrations on startup.

## Rollback

If a deployment contains bugs or issues, you can roll back to a previous stable tag or git SHA without modifying the `docker-compose.yml` file:

1. Open `.env` and set `APP_VERSION` to the target tag (e.g., `v1.0.2` or a specific git SHA like `3d9e81f`):
   ```env
   APP_VERSION=v1.0.2
   ```
2. Pull the targeted version and restart the services:

```bash
docker compose pull
docker compose up -d
```

If you are using a local build setup for development with `docker-compose.override.yml`:

```bash
git pull
docker compose build --pull
docker compose up -d
```

## Troubleshooting

- **Caddy certificate fails:** confirm `DOMAIN` is only the hostname, `BASE_URL` includes `https://`, DNS points to this VM, ports `80` and `443` are open, and Cloudflare proxy is disabled.
- **GHCR pull fails/denied:** verify the VM has internet access and can reach `ghcr.io`. If you cannot pull the prebuilt image, you can fall back to local builds by copying the override configuration (`cp docker-compose.override.yml.example docker-compose.override.yml`) and building from source: `docker compose build`.
- **Litestream credential errors:** verify all `LITESTREAM_R2_*` values in `.env`.
- **App unhealthy:** check `docker compose logs app`; the Docker healthcheck calls `/healthz`, which pings the SQLite connection.
- **Permission denied running Docker:** log out and back in after `sudo usermod -aG docker "$USER"`, or temporarily run Docker commands with `sudo`.

## Production Checklist

- [ ] `DOMAIN` is a bare hostname.
- [ ] `BASE_URL` is the matching `https://...` URL.
- [ ] `SESSION_SECRET` is random and private.
- [ ] DNS points at the Oracle VM.
- [ ] Ports `80` and `443` are open.
- [ ] `docker compose ps` shows `app` healthy and all services running.
- [ ] Litestream logs show successful replication.
- [ ] A restore test has been performed before real data is entered.
