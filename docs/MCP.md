# MCP Integration Guide

This project is designed to be deployed and maintained using three MCP servers in Cursor.

## Required MCPs

| MCP | Purpose |
|-----|---------|
| **Render MCP** | Backend deploy, env vars, logs, PostgreSQL |
| **Vercel MCP** | Frontend deploy, env vars, redeploy |
| **MasterFabric Academy MCP** | Mentor personas (`staff-engineer`, `security-coach`) and academy skills |

## Cursor setup

1. Open **Cursor Settings → MCP**
2. Add each server (or enable if preconfigured)
3. Authenticate when prompted (`mcp_auth`)

### Render MCP (backend)

Use for:

- Set `DATABASE_URL`, `JWT_SECRET`, `ALLOWED_ORIGINS`
- Trigger manual deploy after `main` push
- Read build logs (`./server` start errors)

Example prompts:

- "List my Render services and show the latest deploy status for mlc-llm-monitoring"
- "Set ALLOWED_ORIGINS on my Render backend to https://mlc-llm-monitoring.vercel.app"

### Vercel MCP (frontend)

Use for:

- Set `NEXT_PUBLIC_API_URL` to your Render backend URL
- Redeploy production after env changes
- Rename project to `app-review-monitoring`

Example prompts:

- "Set NEXT_PUBLIC_API_URL on Vercel to https://my-api.onrender.com and redeploy"
- "Show latest Vercel deployment status for mlc-llm-monitoring"

### MasterFabric Academy MCP

Use for:

- `get_mentor_persona` → load **staff-engineer** and **security-coach**
- `get_academy_skill` → academy best practices

Example prompts:

- "Load staff-engineer and security-coach personas and review my auth endpoints"
- "Use security-coach to audit JWT and CORS configuration"

## Mentor personas (reference)

When MCP is connected, load these before code review or deployment:

### staff-engineer

- Production-ready defaults (env vars, migrations, graceful errors)
- Minimal scope, clear naming, no over-engineering
- Deployment checklist: build → start command match, health check, CORS

### security-coach

- Auth: bcrypt passwords, JWT secret from env, token on protected routes
- CORS: explicit `ALLOWED_ORIGINS` in production
- User data isolation: all queries scoped by `user_id`
- No secrets in git; validate email format on register

## Deployment workflow (MCP-driven)

```
1. [Academy MCP]  Load staff-engineer + security-coach personas
2. [Git]          Push to main
3. [Render MCP]   Verify backend deploy + /health
4. [Vercel MCP]   Set NEXT_PUBLIC_API_URL + redeploy
5. [Academy MCP]  Security review of live config
```

## Troubleshooting MCP

| Issue | Action |
|-------|--------|
| Academy MCP timeout | Re-authenticate via `mcp_auth` in Cursor |
| Render/Vercel not listed | Add server in Cursor MCP settings |
| Env not applied on Vercel | Redeploy after saving variables |
