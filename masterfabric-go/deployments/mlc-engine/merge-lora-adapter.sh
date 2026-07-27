#!/bin/bash
set -euo pipefail

# Merge a PEFT LoRA adapter into the base model and quantize for MLC serve.
# Usage: merge-lora-adapter.sh <adapter-name> <lora-path-under-/adapters> [base-model-hf-id]
#
# Example:
#   merge-lora-adapter.sh review-lora raw/review-lora
#   merge-lora-adapter.sh review-lora raw/review-lora google/gemma-2-2b-it

NAME="${1:?adapter name required (e.g. review-lora)}"
LORA_REL="${2:?lora path required (e.g. raw/review-lora)}"
BASE_MODEL="${3:-${MLC_BASE_MODEL:-google/gemma-2-2b-it}}"
QUANT="${MLC_QUANTIZATION:-q4f16_1}"
CONV="${MLC_CONV_TEMPLATE:-gemma_instruction}"
MODEL_TYPE="${MLC_MODEL_TYPE:-gemma2}"

LORA_PATH="/adapters/${LORA_REL#/adapters/}"
OUT="/adapters/merged/${NAME}"
BASE_CACHE="/models/base/${BASE_MODEL//\//--}"

if [[ ! -d "$LORA_PATH" ]]; then
  echo "LoRA adapter not found: $LORA_PATH" >&2
  echo "Drop a HuggingFace PEFT directory under deployments/peft-adapters/raw/" >&2
  exit 1
fi

mkdir -p "$OUT" "$BASE_CACHE"

if [[ ! -f "$BASE_CACHE/config.json" ]]; then
  echo "Downloading base model ${BASE_MODEL} to ${BASE_CACHE}..."
  python3 - <<PY
from huggingface_hub import snapshot_download
snapshot_download(
    repo_id="${BASE_MODEL}",
    local_dir="${BASE_CACHE}",
    local_dir_use_symlinks=False,
)
PY
fi

echo "Merging LoRA ${LORA_PATH} into ${BASE_MODEL} → ${OUT} (${QUANT})..."
python3 -m mlc_llm convert_weight "$BASE_CACHE" \
  --model-type "$MODEL_TYPE" \
  --quantization "$QUANT" \
  --source "$BASE_CACHE" \
  --source-format huggingface-safetensor \
  --output "$OUT" \
  --lora-adapter "$LORA_PATH" \
  --device "${MLC_CONVERT_DEVICE:-cuda}"

python3 -m mlc_llm gen_config "$BASE_CACHE" \
  --model-type "$MODEL_TYPE" \
  --quantization "$QUANT" \
  --conv-template "$CONV" \
  --output "$OUT"

MODEL_ID="${NAME}-${QUANT}-MLC"
if [[ -f "$OUT/mlc-chat-config.json" ]]; then
  python3 - <<PY
import json
from pathlib import Path
path = Path("${OUT}") / "mlc-chat-config.json"
data = json.loads(path.read_text())
data["model_id"] = "${MODEL_ID}"
path.write_text(json.dumps(data, indent=2) + "\n")
PY
fi

echo ""
echo "Done. Merged MLC weights: ${OUT}"
echo "Set in .env.hybrid:"
echo "  MLC_LORA_ADAPTER=${NAME}"
echo "Then recreate mlc-engine. Use model id: ${MODEL_ID}"
