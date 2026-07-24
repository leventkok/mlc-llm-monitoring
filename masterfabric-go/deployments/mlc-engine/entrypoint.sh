#!/bin/bash
set -euo pipefail

export MAX_JOBS="${MAX_JOBS:-4}"
export FLASHINFER_CUDA_ARCH_LIST="${FLASHINFER_CUDA_ARCH_LIST:-8.6}"

exec python3 -m mlc_llm serve "$@"
