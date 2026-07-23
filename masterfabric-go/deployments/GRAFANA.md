# KPI metrics & Grafana

Prometheus KPIs are exposed at **`GET /metrics`** (no auth).  
Business stats for the Monitoring UI are at **`GET /stats`** (auth required).

## Hybrid (Render API + local MLC)

1. Start hybrid stack (rebuild after code changes):

```bash
cd masterfabric-go/deployments
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid up --build -d
```

2. On Render, set **`METRICS_ENABLED=true`** (or redeploy after `render.yaml` push).

3. Open Grafana: http://127.0.0.1:3001 (`admin` / `admin`)

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

## Public Grafana (Cloudflare)

Add a second tunnel route: `grafana.yourdomain.com` → `http://grafana:3000`.  
Change `GRAFANA_ADMIN_PASSWORD` from the default before exposing publicly.
