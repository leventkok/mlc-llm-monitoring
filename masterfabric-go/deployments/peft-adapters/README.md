# PEFT / LoRA adapters

Drop HuggingFace PEFT LoRA directories under `raw/`, merge them into MLC weights under `merged/`, then point `MLC_LORA_ADAPTER` at the merged name.

MLC 0.20 merges LoRA into the base model at **convert time** (no runtime hot-swap). Switch adapters by changing `MLC_LORA_ADAPTER` and restarting `mlc-engine`.

## Layout

```
peft-adapters/
  raw/       ← HuggingFace LoRA adapter dirs (adapter_config.json + weights)
  merged/    ← MLC-quantized output (generated; do not commit large binaries)
```

## Merge an adapter

1. Copy or clone a LoRA trained on **gemma-2-2b-it** into `raw/my-adapter/`.
2. From `masterfabric-go/deployments`:

   **Windows**

   ```cmd
   merge-lora-adapter.cmd my-adapter raw\my-adapter
   ```

   **Linux / macOS**

   ```bash
   ./merge-lora-adapter.sh my-adapter raw/my-adapter
   ```

3. Set in `.env.hybrid`:

   ```env
   MLC_LORA_ADAPTER=my-adapter
   ```

4. Restart the engine:

   ```bash
   docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml \
     --env-file .env.hybrid up -d --force-recreate mlc-engine mlc-llm
   ```

5. Update Render `MLC_LLM_MODEL` to match the served model id (shown in merge script output / `/v1/models`).

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `MLC_BASE_MODEL` | `google/gemma-2-2b-it` | HF base for LoRA merge |
| `MLC_LORA_ADAPTER` | *(empty)* | Serve `/adapters/merged/<name>` instead of `MLC_MODEL` |
| `MLC_QUANTIZATION` | `q4f16_1` | Quantization for merged weights |
| `MLC_CONV_TEMPLATE` | `gemma_instruction` | Chat template for `gen_config` |
