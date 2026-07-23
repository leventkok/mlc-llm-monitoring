@echo off
cd /d "%~dp0"
if not exist .env.hybrid (
  copy .env.hybrid.example .env.hybrid
  echo Created .env.hybrid — edit MLC_API_KEY and CLOUDFLARE_TUNNEL_TOKEN before production use.
)
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid --profile tunnel up --build
