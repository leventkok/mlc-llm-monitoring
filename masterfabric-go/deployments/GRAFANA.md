# KPI metrics & Grafana

Prometheus KPIs are exposed at **`GET /metrics`** (no auth).  
Business stats for the Monitoring UI are at **`GET /stats`** (auth required).

Grafana is **admin-only in the product UI** (link on `/admin`). The public hostname still needs its own lock (Grafana login + Cloudflare Access).

## Hybrid (Render API + local MLC)

1. Start hybrid stack (rebuild after code changes):

```bash
cd masterfabric-go/deployments
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid up --build -d
```

2. On Render, set **`METRICS_ENABLED=true`** (or redeploy after `render.yaml` push).

3. Open Grafana locally: http://127.0.0.1:3001 (use `GRAFANA_ADMIN_USER` / `GRAFANA_ADMIN_PASSWORD` from `.env.hybrid`)

4. Dashboard: **LLM Monitoring → LLM Review Monitoring KPIs**

Prometheus scrapes:

| Job | Target |
|-----|--------|
| `render-api` | `https://mlc-llm-monitoring.onrender.com/metrics` |
| `mlc-llm` | local Docker mock |
| `prometheus` | self |

Verify targets: http://127.0.0.1:9090/targets — `render-api` should be **UP** after Render deploy.

## KPI metrics

| Metric | Description |
|--------|-------------|
| `app_review_llm_analyze_total` | Analyze success/error count |
| `app_review_llm_inference_duration_seconds` | Inference latency histogram |
| `app_review_decisions_total` | Decisions by category + sentiment |
| `app_review_auto_score_quality` | Auto quality score (1–5) |
| `app_review_reviews_created_total` | Reviews created |
| `app_review_http_requests_total` | HTTP traffic |
| `mlc_inference_requests_total` | Local MLC mock calls |

## Public Grafana (Cloudflare Tunnel)

Tunnel route: `grafana.inferreview.com` → `http://grafana:3000` (same connector as MLC).

### Layer 1 — Grafana login (required)

In `.env.hybrid` / `.env.prod`:

```env
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=<strong-random-password>
GRAFANA_ROOT_URL=https://grafana.inferreview.com
GRAFANA_COOKIE_SECURE=true
```

Compose sets `GF_AUTH_ANONYMOUS_ENABLED=false` and `GF_USERS_ALLOW_SIGN_UP=false`.

Restart Grafana after changing env:

```bash
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid up -d grafana
```

### Layer 2 — Cloudflare Access (recommended)

Hiding the link in the app is not enough: anyone who knows the URL can still hit the login page. Add **Zero Trust Access** in front of the tunnel hostname.

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → **Access** → **Applications** → **Add an application**
2. Type: **Self-hosted**
3. Application domain: `grafana.inferreview.com` (same zone as your tunnel)
4. **Add a policy** → Action: **Allow**
5. Include rule: **Emails** → your admin addresses (e.g. `levent.kok.dev@gmail.com`)
6. Save

Visitors must pass Cloudflare Access (email OTP / IdP) **before** Grafana’s login form.

Optional: reuse the same policy for `mlc.inferreview.com` if you want MLC API hidden from the open internet (Render already uses `MLC_API_KEY`).

### Layer 3 — Local bind only (optional)

`GRAFANA_HOST_BIND=127.0.0.1` keeps Grafana off the LAN; only the tunnel (or SSH port-forward) reaches it. Default in compose files.

## Checklist

| Step | Done? |
|------|-------|
| Strong `GRAFANA_ADMIN_PASSWORD` (not `admin`) | |
| `GRAFANA_ROOT_URL` matches public URL | |
| Cloudflare Access on `grafana.inferreview.com` | |
| Product link only on `/admin` (not Monitoring) | |
