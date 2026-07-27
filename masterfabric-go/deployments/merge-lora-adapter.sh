#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <adapter-name> <raw-subdir> [base-model-hf-id]" >&2
  echo "Example: $0 review-lora raw/review-lora" >&2
  exit 1
fi

docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml \
  --env-file .env.hybrid run --rm --gpus all \
  --entrypoint /usr/local/bin/merge-lora-adapter.sh \
  mlc-engine "$@"
