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

## App rename

Display name is **app-review-monitoring**. To rename the Vercel project URL, update the project name in the Vercel dashboard.
