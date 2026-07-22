# Local Docker Stack

Run the full local stack (Postgres, backend, MLC mock, Nginx gateway, Prometheus, Grafana).

## Prerequisites

- Docker Desktop (or Docker Engine + Compose v2)
- Ports free: `8080` (API gateway), `9090` (Prometheus), `3001` (Grafana)

## Quick start

```bash
cd masterfabric-go/deployments
cp .env.docker.example .env.docker
docker compose -f docker-compose.stack.yml --env-file .env.docker up --build
```

## Endpoints

| Service | URL |
|---------|-----|
| API gateway | http://127.0.0.1:8080 |
| Health | http://127.0.0.1:8080/health |
| Metrics | http://127.0.0.1:8080/metrics |
| Prometheus | http://127.0.0.1:9090 |
| Grafana | http://127.0.0.1:3001 (admin / admin) |
| MLC mock (via gateway) | http://127.0.0.1:8080/llm/health |

## Frontend (local)

Point the Next.js app at the gateway:

```env
# frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_USE_SERVER_INFERENCE=true
```

Then:

```bash
cd frontend
npm run dev
```

Open http://localhost:3000 — Analyze uses Docker MLC (`POST /reviews/{id}/analyze`) instead of WebLLM.

To use browser WebLLM again, remove `NEXT_PUBLIC_USE_SERVER_INFERENCE` or set it to `false`.

## Horizontal scale (local)

```bash
docker compose -f docker-compose.stack.yml --env-file .env.docker up --build --scale backend=2 --scale mlc-llm=2
```

Nginx load-balances `backend` replicas via Docker Compose DNS.

## Replace MLC mock with real MLC LLM

The `mlc-llm` service is a lightweight OpenAI-compatible mock for local pipelines.
When you have GPU + a real MLC server image, replace the `mlc-llm` build section in
`docker-compose.stack.yml` and keep `MLC_LLM_BASE_URL=http://mlc-llm:8080`.

## Stop

```bash
docker compose -f docker-compose.stack.yml --env-file .env.docker down
```

Remove volumes:

```bash
docker compose -f docker-compose.stack.yml --env-file .env.docker down -v
```
