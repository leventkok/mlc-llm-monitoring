# app-review-monitoring

Raw LLM monitoring and decision scoring for app store reviews. Frontend runs Gemma in the browser via Web MLC; backend handles auth and per-user review data.

## Local development

### Backend (Go + PostgreSQL)

```bash
# .env in repo root
DATABASE_URL=postgres://user:pass@localhost:5432/app_review_monitoring
JWT_SECRET=your-secret
PORT=8080
```

```bash
go run .
```

### Frontend (Next.js)

```bash
cd frontend
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
```

## Production deployment

### Render (backend)

**Build Command** (Settings → Build & Deploy):

```
go build -tags netgo -ldflags '-s -w' -o server .
```

**Start Command**:

```
./server
```

Build ve Start komutları **aynı dosya adını** kullanmalı. Render varsayılanı `-o app` / `./app` üretir; Start `./server` ise build de `-o server` olmalı.

Set environment variables:

- `DATABASE_URL` — PostgreSQL connection string
- `JWT_SECRET` — signing secret for auth tokens
- `ALLOWED_ORIGINS` — e.g. `https://mlc-llm-monitoring.vercel.app,https://app-review-monitoring.vercel.app`

### Vercel (frontend)

Set environment variable:

- `NEXT_PUBLIC_API_URL` — your Render backend URL, e.g. `https://your-app.onrender.com`

**Important:** Without `NEXT_PUBLIC_API_URL` pointing to the live backend, remote users will see "failed to fetch" because the app defaults to `http://localhost:8080`.

## Auth

- Register with **email**, **username**, and **password**
- Sign in with **email** and **password**
- Each user sees only their own reviews, decisions, and metrics
- Delete account from **Settings** (`/settings`)

## API

20 endpoints — see [docs/ENDPOINTS.md](docs/ENDPOINTS.md)

| Group | Count |
|-------|-------|
| Common | 1 |
| Config | 2 |
| Auth | 9 |
| WEB MLC-LLM | 8 |

## MCP integration

Deploy and review using **Render MCP**, **Vercel MCP**, and **MasterFabric Academy MCP**. See [docs/MCP.md](docs/MCP.md).

Load mentor personas before reviews:

- `get_mentor_persona` → **staff-engineer**
- `get_mentor_persona` → **security-coach**

## App rename

Display name is **app-review-monitoring**. To rename the Vercel URL:

1. Vercel Dashboard → Project → **Settings → General**
2. **Project Name** → `app-review-monitoring`
3. New URL: `https://app-review-monitoring.vercel.app`
4. Update Render `ALLOWED_ORIGINS` with the new URL
