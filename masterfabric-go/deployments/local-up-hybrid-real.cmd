@echo off
cd /d "%~dp0"
if not exist .env.hybrid (
  copy .env.hybrid.example .env.hybrid
  echo Created .env.hybrid — edit MLC_API_KEY and CLOUDFLARE_TUNNEL_TOKEN before production use.
)
echo Starting hybrid stack with REAL MLC (GPU). First run may take 10+ minutes to download the model.
docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml --env-file .env.hybrid --profile tunnel up --build
