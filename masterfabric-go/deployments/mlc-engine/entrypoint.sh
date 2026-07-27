#!/bin/bash
set -euo pipefail

export MAX_JOBS="${MAX_JOBS:-4}"
export FLASHINFER_CUDA_ARCH_LIST="${FLASHINFER_CUDA_ARCH_LIST:-8.6}"

args=("$@")
if [[ -n "${MLC_LORA_ADAPTER:-}" ]]; then
  adapter_path="/adapters/merged/${MLC_LORA_ADAPTER}"
  if [[ ! -d "$adapter_path" ]]; then
    echo "mlc-engine: MLC_LORA_ADAPTER=${MLC_LORA_ADAPTER} not found at ${adapter_path}" >&2
    echo "Run merge-lora-adapter from deployments/ first." >&2
    exit 1
  fi
  args=("$adapter_path" "${args[@]:1}")
  echo "mlc-engine: serving merged LoRA adapter ${MLC_LORA_ADAPTER} from ${adapter_path}"
fi

exec python3 -m mlc_llm serve "${args[@]}"