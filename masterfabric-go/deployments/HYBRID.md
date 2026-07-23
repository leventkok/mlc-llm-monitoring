# Hybrid Deploy: Render API + Local MLC

Public traffic stays on **Render + Vercel**. Only LLM inference runs on **your machine** in Docker. Render calls your MLC through a **Cloudflare Tunnel** (reverse proxy).

```
User → Vercel → Render (auth, DB, /reviews/{id}/analyze)
                      │
                      │  HTTPS + X-MLC-API-Key
                      ▼
              Cloudflare Tunnel
                      │
                      ▼
              Your PC/server (Docker)
              mlc-gateway (nginx) → mlc-llm
              Grafana + Prometheus (local monitoring, optional)
```

## Prerequisites

- Docker Desktop running
- [Cloudflare](https://dash.cloudflare.com) account (free)
- A domain on Cloudflare (or use a `*.trycloudflare.com` quick tunnel for testing)
- Render service already deployed
- Machine stays **on and online** while prod analyze is used

## Step 1 — Local hybrid stack

```bash
cd masterfabric-go/deployments
cp .env.hybrid.example .env.hybrid
```

Edit `.env.hybrid`:

```env
MLC_API_KEY=<generate-a-long-random-secret>
CLOUDFLARE_TUNNEL_TOKEN=<from Cloudflare dashboard>
```

Generate a key (PowerShell):

```powershell
-join ((48..57 + 65..90 + 97..122) | Get-Random -Count 48 | ForEach-Object {[char]$_})
```

Start stack **with tunnel**:

```bash
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid --profile tunnel up --build
```

Local checks:

| URL | Expected |
|-----|----------|
| http://127.0.0.1:8787/health | `{"status":"ok"}` |
| http://127.0.0.1:3001 | Grafana (admin / admin) |

## Step 2 — Cloudflare Tunnel (inferreview.com)

1. Zero Trust → **Free plan** → **Networks** → **Tunnels** → **Create a tunnel**
2. Type: **Cloudflared** (not Mesh)
3. Name: `inferreview-hybrid`
4. Install connector → **Docker** → copy token into `.env.hybrid` as `CLOUDFLARE_TUNNEL_TOKEN`
   - Do **not** run the standalone `docker run cloudflare/cloudflared…` command; use docker compose below
5. **Public Hostname** tab → add routes:

| Hostname | Type | URL (Docker service) |
|----------|------|----------------------|
| `mlc.inferreview.com` | HTTP | `mlc-gateway:80` |
| `grafana.inferreview.com` | HTTP | `grafana:3000` |

6. Start/restart compose with `--profile tunnel`

Test from any browser:

```
https://mlc.inferreview.com/health
https://grafana.inferreview.com
```

Should return `{"status":"ok"}` and Grafana login respectively.

### Optional: app + API on same domain (DNS only)

| Subdomain | Points to |
|-----------|-----------|
| `app.inferreview.com` | Vercel (custom domain) |
| `api.inferreview.com` | Render (custom domain) |

Update Render `ALLOWED_ORIGINS=https://app.inferreview.com` when frontend moves.

### Quick test without custom domain (ngrok alternative)

If you cannot use Cloudflare yet, expose port 8787 temporarily:

```bash
ngrok http 8787
```

Use the `https://….ngrok-free.app` URL as `MLC_LLM_BASE_URL` on Render (URL changes each restart).

## Step 3 — Render env vars

Render Dashboard → **app-review-monitoring-api** → **Environment**:

| Key | Value |
|-----|-------|
| `MLC_LLM_ENABLED` | `true` |
| `MLC_LLM_BASE_URL` | `https://mlc.inferreview.com` |
| `MLC_LLM_API_KEY` | same as `MLC_API_KEY` in `.env.hybrid` |
| `MLC_LLM_MODEL` | `gemma-2-2b-it-q4f16_1-MLC` (optional) |

Save → **Manual Deploy** (or wait for auto deploy after git push).

Verify Render logs contain:

```
server-side mlc inference enabled
```

## Step 4 — Vercel env vars

Vercel → Project → **Settings** → **Environment Variables**:

| Key | Value |
|-----|-------|
| `NEXT_PUBLIC_USE_SERVER_INFERENCE` | `true` |

Redeploy frontend. Prod dashboard will call `POST /reviews/{id}/analyze` on Render instead of WebLLM in the browser.

## Step 5 — End-to-end test

1. Open https://mlc-llm-monitoring.vercel.app
2. Register / login
3. Add a review: `I love this app, it's great`
4. Click **Analyze**
5. Expected: `praise / positive` (mock rules until real MLC GPU image is used)

Watch Grafana at https://grafana.inferreview.com while testing (or http://127.0.0.1:3001 locally).

## Architecture notes

| Component | Where | Public? |
|-----------|-------|---------|
| Frontend | Vercel | Yes |
| API + Postgres | Render | Yes |
| MLC inference | Your Docker | No (tunnel only) |
| Grafana/Prometheus | Your Docker | Tunnel: `grafana.inferreview.com` |

- **Login, reviews, DB** → always Render
- **Analyze only** → Render forwards to your MLC
- **Grafana** → for you only; users never see it

## Replace mock with real MLC

In `docker-compose.hybrid.yml`, replace the `mlc-llm` build section with your GPU MLC image. Keep `MLC_LLM_BASE_URL` pointing at the tunnel hostname.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Analyze returns 503 | Render `MLC_LLM_ENABLED` / `MLC_LLM_BASE_URL` wrong |
| 403 from MLC | `MLC_LLM_API_KEY` on Render ≠ `MLC_API_KEY` locally |
| Tunnel down | Docker + cloudflared running? Token valid? |
| Analyze timeout | Render free tier cold start + tunnel latency; retry |
| Works locally, not prod | Vercel missing `NEXT_PUBLIC_USE_SERVER_INFERENCE=true` |

## Windows one-liner

```cmd
local-up-hybrid.cmd
```

(Ensures `.env.hybrid` exists, starts stack + tunnel profile.)
