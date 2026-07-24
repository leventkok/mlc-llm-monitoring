# Real MLC Docker (GPU)

Replace the keyword-based **mlc-mock** with the official **MLC LLM** image (`gemma-2-2b-it-q4f16_1-MLC`) while keeping the same public tunnel URL and Render integration.

```
Render → Cloudflare Tunnel → mlc-gateway (nginx) → mlc-llm (auth proxy) → mlc-engine (GPU)
```

Mock mode (default) is unchanged — only add the override file when you have a GPU ready.

## Architecture

| Service | Role |
|---------|------|
| `mlc-engine` | Builds from `mlc-engine/Dockerfile` — **stable** wheels `mlc-ai-cu130==0.20.0` + `mlc-llm-cu130==0.20.0.dev0` from [mlc.ai/wheels](https://mlc.ai/wheels) |
| `mlc-llm` | **mlc-proxy** — validates `X-MLC-API-Key`, forwards `/v1/*`, exposes `/health` + `/metrics` |
| `mlc-gateway` | Same nginx config as mock stack |
| `cloudflared` | Same tunnel profile |

Render env vars (`MLC_LLM_BASE_URL`, `MLC_LLM_API_KEY`, `MLC_LLM_MODEL`) do **not** change.

## Prerequisites

1. **NVIDIA GPU** with recent drivers
2. **Docker Desktop** (Windows) with WSL2 backend, or Linux
3. **NVIDIA Container Toolkit** — [install guide](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)
4. Verify GPU in Docker:

   ```bash
   docker run --rm --gpus all nvidia/cuda:12.0.0-base-ubuntu22.04 nvidia-smi
   ```

5. Enough disk for the model cache (~2 GB); stored in Docker volume `mlc_models`

## Configuration

Add to `.env.hybrid` (optional — defaults are fine):

```env
MLC_MODEL=HF://mlc-ai/gemma-2-2b-it-q4f16_1-MLC
MLC_DEVICE=cuda
MLC_WHEEL_SUFFIX=cu130
```

Default model is **gemma-2-2b-it-q4f16_1-MLC** (same as mock stack / Render). First startup downloads ~2 GB and JIT-compiles on GPU (10–20 min). `MLC_WHEEL_SUFFIX` must match your CUDA driver (see [mlc.ai/wheels](https://mlc.ai/wheels)).

`MLC_API_KEY` and `CLOUDFLARE_TUNNEL_TOKEN` are required (same as mock hybrid).

## Start (real MLC + tunnel)

```bash
cd masterfabric-go/deployments
docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml \
  --env-file .env.hybrid --profile tunnel up --build
```

Windows:

```cmd
local-up-hybrid-real.cmd
```

**First startup:** Docker builds `mlc-engine` (pip install, several minutes), then downloads the HuggingFace model. Health checks allow up to ~15 minutes (`start_period: 900s`). Watch logs:

```bash
docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml logs -f mlc-engine
```

When ready:

| URL | Expected |
|-----|----------|
| http://127.0.0.1:8787/health | `{"status":"ok"}` |
| https://mlc.inferreview.com/health | `{"status":"ok"}` |

## Test inference locally

```bash
curl -s -X POST http://127.0.0.1:8787/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-MLC-API-Key: YOUR_MLC_API_KEY" \
  -d '{"model":"gemma-2-2b-it-q4f16_1-MLC","messages":[{"role":"user","content":"Classify: love this app"}],"max_tokens":60}'
```

Then run **Analyze** on the Vercel dashboard — Render still calls the same tunnel URL.

## Roll back to mock

Stop the stack and start without the override file:

```bash
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid --profile tunnel up --build
```

Or use `local-up-hybrid.cmd`.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `pull access denied for mlcaidev/mlc-llm` | Fixed — we build `mlc-engine` locally from pip wheels, not Docker Hub |
| `mlc-engine` exits immediately | GPU not visible in Docker — check `nvidia-smi` in a `--gpus all` container |
| pip wheel / CUDA mismatch | Try `MLC_WHEEL_SUFFIX=cu128` or `cu130` in `.env.hybrid` and rebuild |
| `No module named 'tvm'` | Rebuild `mlc-engine` — Dockerfile pins matching mlc-ai + mlc-llm wheel versions |
| `/usr/local/cuda/bin/nvcc: not found` | Fixed — use `cuda:*-devel` base image (not `runtime`); FlashInfer JIT needs nvcc |
| Nightly compile errors (`CallTIRRewrite`, `DataflowVar`) | Use stable wheels in Dockerfile (default since 0.20.0); avoid `mlc-*-nightly` unless you accept ~30 GB images |
| `DataflowVar` / FlashInfer errors | Nightly-only; stable 0.20 uses runtime CUDA image, no FlashInfer JIT |
| Health stays `starting` for a long time | Normal on first run (model download + JIT compile, 10–20 min). Check `mlc-engine` logs |
| 403 from MLC | `MLC_LLM_API_KEY` on Render ≠ `MLC_API_KEY` in `.env.hybrid` |
| Analyze slow | Real GPU inference + tunnel latency; Render timeout is 120s |
| Out of VRAM | Use a smaller quant model or `--device cpu` (very slow; mock is better for dev) |

## PEFT / LoRA adapters

MLC 0.20 merges LoRA at **convert time** (not runtime hot-swap). Adapters live on a bind-mounted volume:

```
deployments/peft-adapters/
  raw/     ← HuggingFace PEFT LoRA dirs
  merged/  ← MLC-quantized output (generated)
```

1. Put a LoRA trained on **gemma-2-2b-it** in `peft-adapters/raw/my-adapter/`.
2. Merge (GPU, several minutes):

   ```cmd
   merge-lora-adapter.cmd my-adapter raw\my-adapter
   ```

3. In `.env.hybrid`:

   ```env
   MLC_LORA_ADAPTER=my-adapter
   ```

4. Recreate the engine:

   ```bash
   docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml \
     --env-file .env.hybrid up -d --force-recreate mlc-engine mlc-llm
   ```

5. Set Render `MLC_LLM_MODEL` to the model id printed by the merge script (check `/v1/models`).

See [peft-adapters/README.md](./peft-adapters/README.md) for env vars and troubleshooting.

## Monitoring

Prometheus scrapes `mlc-llm:8080/metrics` (proxy counters). Grafana dashboards work the same as the mock stack.
