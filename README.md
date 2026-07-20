<div align="center">

<a href="https://academy.masterfabric.co">
  <img src="https://academy.masterfabric.co/academy-badge.png" width="120" alt="MasterFabric Academy">
</a>

<p>
  <sub>
    academy.masterfabric.co is a
    <a href="https://masterfabric.co">MasterFabric</a>
    subsidiary.
  </sub>
</p>

</div>

# app-review-monitoring

**App Review Monitoring** — observe how a raw language model classifies app-store reviews, then score the quality of its decisions.

**Live:** https://mlc-llm-monitoring.vercel.app

---

## What is this project?

A full-stack monitoring dashboard for **raw LLM decision-making**. Users paste or add app-store reviews, run them through **Gemma 2** entirely in the browser (Web MLC-LLM), and then evaluate whether the model's category and sentiment judgments are correct.

The backend does not run the LLM. It handles **auth**, **per-user data storage**, and **metrics aggregation**. This keeps inference local and raw — no server-side model wrapper or post-processing layer.

---

## How it works

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Browser (Next.js on Vercel)                                │
│                                                             │
│  ┌─────────────┐    ┌──────────────────────────────────┐   │
│  │ Auth views  │    │ Master views                      │   │
│  │ login       │    │ Home → Dashboard → Monitoring     │   │
│  │ register    │    │         ↘ Settings                │   │
│  └──────┬──────┘    └──────────────┬───────────────────┘   │
│         │                           │                       │
│         │         ┌─────────────────▼─────────────────┐     │
│         │         │  @mlc-ai/web-llm (Gemma 2 2B)      │     │
│         │         │  Runs locally in the browser       │     │
│         │         │  → category + sentiment + latency  │     │
│         │         └─────────────────┬─────────────────┘     │
│         │                           │ decisions             │
└─────────┼───────────────────────────┼───────────────────────┘
          │ JWT                       │
          ▼                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Go API (Render) + PostgreSQL                                 │
│  Auth · Reviews · Decisions · Scores · Metrics (per user)    │
└─────────────────────────────────────────────────────────────┘
```

### User flow

1. **Register / sign in** with email, username, and password.
2. **Dashboard** — add a review (app name, store, rating, text).
3. Click **Analyze** — Gemma loads in the browser (cached after first run) and returns a raw JSON verdict: category (`bug`, `feature`, `praise`, `spam`, `other`) and sentiment (`positive`, `negative`, `neutral`).
4. The decision is saved to the backend, scoped to the logged-in user.
5. **Monitoring** — review all decisions, rate quality (1–5), optionally mark the correct category, and watch accuracy metrics update.
6. **Settings** — change password or delete the account (and all associated data).

Each user sees only their own reviews, decisions, and metrics.

---

## Application structure

| View | Route | Role |
|------|-------|------|
| Sign in | `/login` | Auth |
| Register | `/register` | Auth |
| Home | `/` | Master — entry point after login |
| Dashboard | `/dashboard` | Master — add reviews, trigger Gemma analysis |
| Monitoring | `/monitoring` | Master — score decisions, view accuracy |
| Settings | `/settings` | Subview — password change, account deletion |

The frontend is a **Next.js SPA** with client-side routing and protected routes. Unauthenticated users are redirected to `/login`.

---

## LLM approach

| Aspect | Choice |
|--------|--------|
| Model | `gemma-2-2b-it-q4f16_1-MLC` |
| Runtime | [@mlc-ai/web-llm](https://www.npmjs.com/package/@mlc-ai/web-llm) in the browser |
| Inference | Client-side only — no backend `/analyze` mock or server proxy |
| Output | Raw model JSON — category, sentiment, latency ms |
| Scoring | Human-in-the-loop on the Monitoring page |

This satisfies the **raw LLM monitoring + decision scoring** base case: the model decides; humans observe and score.

---

## Backend design

Go REST API with **20 endpoints**:

| Group | Count | Responsibility |
|-------|------:|----------------|
| Common | 1 | Health check |
| Config | 2 | App name, model id, version |
| Auth | 9 | Register, login, JWT, profile, password, account delete |
| WEB MLC-LLM | 8 | Reviews, decisions, scores, metrics (all user-scoped) |

Full endpoint list: [docs/ENDPOINTS.md](docs/ENDPOINTS.md)

**Stack:** Go · net/http · PostgreSQL (pgx) · JWT · bcrypt

---

## Tech stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js 16, React 19, Tailwind CSS 4 |
| LLM | Web MLC-LLM, Gemma 2 2B |
| Backend | Go, PostgreSQL |
| Frontend deploy | [Vercel](https://mlc-llm-monitoring.vercel.app) |
| Backend deploy | [Render](https://render.com) |

---

## Getting started (local)

### 1. Backend

```bash
# .env in repo root
DATABASE_URL=postgres://user:pass@localhost:5432/app_review_monitoring
JWT_SECRET=your-secret
PORT=8080
ALLOWED_ORIGINS=http://localhost:3000
```

```bash
go run .
```

### 2. Frontend

```bash
cd frontend
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
```

Open http://localhost:3000

---

## Production deployment

### Render (backend)

| Setting | Value |
|---------|-------|
| Build Command | `go build -tags netgo -ldflags '-s -w' -o server .` |
| Start Command | `./server` |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Signing secret |
| `ALLOWED_ORIGINS` | `https://mlc-llm-monitoring.vercel.app` |

See [render.yaml](render.yaml) for Blueprint config.

### Vercel (frontend)

| Setting | Value |
|---------|-------|
| Root Directory | `frontend` |
| `NEXT_PUBLIC_API_URL` | Render backend URL |

Redeploy after env changes — variables are embedded at build time.

---

## MCP tooling

Development and deployment use three MCP servers in Cursor:

| MCP | Use |
|-----|-----|
| **Render MCP** | Backend deploy, env vars, logs |
| **Vercel MCP** | Frontend deploy, env vars |
| **MasterFabric Academy MCP** | Mentor personas (`staff-engineer`, `security-coach`), academy skill |

Setup guide: [docs/MCP-SETUP-TR.md](docs/MCP-SETUP-TR.md)

---

## Project scope (MasterFabric Academy)

| Requirement | Implementation |
|-------------|----------------|
| Next.js SPA, ≥3 master views + auth | Home, Dashboard, Monitoring, Settings + login/register |
| Web MLC-LLM (Gemma) on Vercel | Client-side Gemma via `@mlc-ai/web-llm` |
| Go backend ≥20 endpoints | 20 EP — Config[2] + Auth[9] + WEB MLC-LLM[8] + Common[1] |
| Render + Vercel live | Production deployed |
| Raw LLM monitoring + decision scoring | Dashboard analyze + Monitoring score |
| MCP usage | Render, Vercel, Academy MCP |

---

## Links

- **Live app:** https://mlc-llm-monitoring.vercel.app
- **Repository:** https://github.com/leventkok/mlc-llm-monitoring
- **Academy:** https://academy.masterfabric.co
