@echo off
setlocal
cd /d "%~dp0"

if not exist .env.docker (
  copy .env.docker.example .env.docker
)

docker compose -f docker-compose.stack.yml --env-file .env.docker up --build %*
