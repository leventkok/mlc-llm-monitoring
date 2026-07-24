@echo off
setlocal
cd /d "%~dp0"

if "%~1"=="" (
  echo Usage: merge-lora-adapter.cmd ^<adapter-name^> ^<raw-subdir^> [base-model-hf-id]
  echo Example: merge-lora-adapter.cmd review-lora raw\review-lora
  exit /b 1
)

docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml ^
  --env-file .env.hybrid run --rm --gpus all ^
  --entrypoint /usr/local/bin/merge-lora-adapter.sh ^
  mlc-engine %*
