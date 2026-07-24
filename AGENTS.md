# AGENTS.md — App Review Monitoring (inferreview.com)

Guidance for AI agents and contributors working in this repository.

## What this project is

App store review monitoring stack:

| Surface | Role |
|---------|------|
| **frontend/** | Next.js UI — dashboard, rich results, monitoring |
| **masterfabric-go/** | Go API — auth, reviews, LLM analyze, `/stats`, Prometheus `/metrics` |
| **deployments/** | Docker hybrid stack — MLC mock, Grafana, Prometheus, Cloudflare tunnel |

Live prod (do not break casually): Vercel frontend → Render API → tunneled local MLC + Grafana on `inferreview.com`.

## Layout

```
frontend/                 # Next.js app (Vercel)
masterfabric-go/          # Go backend (Render)
  cmd/server/             # Entry point
  internal/               # Clean architecture (domain / application / infrastructure)
  deployments/            # docker-compose, Grafana, MLC mock, HYBRID.md
.cursor/rules/            # Repo-wide agent rules (branch, PR, commits)
```

## Agent workflow

Follow the same delivery model as [masterfabric-mac-cli](https://github.com/gurkanfikretgunak/masterfabric-mac-cli):

1. Read `.cursor/rules/release-workflow.mdc` before committing or opening PRs.
2. Use a **feature/fix branch** — e.g. `feat/rich-result-monitoring`, `fix/grafana-kpi-query`.
3. Write **English** commit messages (Conventional Commits — see rule file).
4. Push the branch and open an **English** PR with **Summary** + **Test plan**.
5. Do **not** change live URLs, tunnel tokens, or deploy config without explicit user approval.
6. Never commit secrets (`.env.hybrid`, API keys, tunnel tokens).
7. Keep PR titles, commit messages, and code comments in **English**. Chat may be Turkish.

## Quick verify

```bash
cd masterfabric-go && go build ./...
cd frontend && npm run build
```
