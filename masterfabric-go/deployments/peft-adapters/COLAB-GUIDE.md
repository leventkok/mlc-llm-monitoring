# Colab’da LoRA eğitimi + MLC convert (disk dostu)

Lokal PC’de **eğitim yok**, **merge yok** — sadece HuggingFace’ten serve.

```
Colab (GPU) → LoRA eğit → MLC convert → HF’ye yükle
Lokal PC    → MLC_MODEL=HF://... → docker compose up
```

## Ön hazırlık (15 dk)

1. **Google Colab:** https://colab.research.google.com → New notebook
2. **Runtime → Change runtime type → T4 GPU**
3. **HuggingFace hesabı:** https://huggingface.co/join
4. **HF Access Token:** Settings → Access Tokens → **Write** yetkili token
5. Repo’da eğitim verisi:
   - Demo: [`sample_train.jsonl`](./sample_train.jsonl) (20 örnek)
   - **Production’a yakın:** [`train_v2.jsonl`](./train_v2.jsonl) (**240 örnek**, backend prompt ile aynı)
   - Yeniden üret: `python generate_train_v2.py`

---

## Adım 1 — Colab’a veri yükle

Colab sol menü **Files** → Upload → `sample_train.jsonl`

İleride 100–500 satır gerçek review eklersen model daha iyi olur.

---

## Adım 2 — Notebook hücreleri (sırayla çalıştır)

### Hücre A — GPU kontrol

```python
!nvidia-smi
```

GPU görünmüyorsa Runtime → T4 seç.

### Hücre B — Paketler

```python
!pip install -q "transformers>=4.43" "peft>=0.11" "trl>=0.9" "datasets" "bitsandbytes" "accelerate" huggingface_hub
```

### Hücre C — HF login

```python
from huggingface_hub import login
login()  # token yapıştır (Write)
```

### Hücre D — Dataset

```python
from datasets import load_dataset

ds = load_dataset("json", data_files="sample_train.jsonl", split="train")
print(ds[0])
```

### Hücre E — Model + QLoRA

```python
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training

BASE = "google/gemma-2-2b-it"
LORA_OUT = "./inferreview-lora"

bnb = BitsAndBytesConfig(
    load_in_4bit=True,
    bnb_4bit_quant_type="nf4",
    bnb_4bit_compute_dtype=torch.bfloat16,
)

tokenizer = AutoTokenizer.from_pretrained(BASE)
model = AutoModelForCausalLM.from_pretrained(
    BASE, quantization_config=bnb, device_map="auto",
)
model = prepare_model_for_kbit_training(model)

lora = LoraConfig(
    r=16, lora_alpha=32, lora_dropout=0.05,
    target_modules=["q_proj", "k_proj", "v_proj", "o_proj",
                    "gate_proj", "up_proj", "down_proj"],
    task_type="CAUSAL_LM",
)
model = get_peft_model(model, lora)
model.print_trainable_parameters()
```

### Hücre F — Eğitim

```python
from trl import SFTTrainer, SFTConfig
from transformers import TrainingArguments

def format_chat(example):
    text = tokenizer.apply_chat_template(
        example["messages"], tokenize=False, add_generation_prompt=False,
    )
    return {"text": text}

train_ds = ds.map(format_chat)

trainer = SFTTrainer(
    model=model,
    train_dataset=train_ds,
    processing_class=tokenizer,
    args=SFTConfig(
        output_dir="./checkpoints",
        num_train_epochs=3,
        per_device_train_batch_size=2,
        gradient_accumulation_steps=4,
        learning_rate=2e-4,
        logging_steps=5,
        save_strategy="no",
        bf16=True,
        max_length=512,
        dataset_text_field="text",
    ),
)
trainer.train()
model.save_pretrained(LORA_OUT)
tokenizer.save_pretrained(LORA_OUT)
print("LoRA saved to", LORA_OUT)
```

~5–15 dk (T4, 20 örnek).

### Hücre G — MLC merge + quantize (Colab’da)

```python
!pip install -q "apache-tvm-ffi==0.1.11"
!pip install -q --no-deps -f https://mlc.ai/wheels "mlc-ai-cu124==0.20.0" "mlc-llm-cu124==0.20.0.dev0"
!pip install -q cloudpickle scipy tornado ml_dtypes numpy packaging typing_extensions \
    attrs datasets fastapi openai pandas prompt_toolkit requests safetensors \
    sentencepiece shortuuid tiktoken torch tqdm transformers uvicorn aiohttp pydantic flashinfer-python
```

`cu124` Colab’da hata verirse `cu121` dene.

```python
import os
MERGED = "./inferreview-merged-mlc"
os.makedirs(MERGED, exist_ok=True)

!python -m mlc_llm convert_weight google/gemma-2-2b-it \
  --model-type gemma2 \
  --quantization q4f16_1 \
  --lora-adapter {LORA_OUT} \
  --output {MERGED}

!python -m mlc_llm gen_config google/gemma-2-2b-it \
  --model-type gemma2 \
  --quantization q4f16_1 \
  --conv-template gemma_instruction \
  --output {MERGED}
```

~10–20 dk.

### Hücre H — HuggingFace’e yükle

```python
from huggingface_hub import HfApi, create_repo

HF_REPO = "YOUR_USERNAME/inferreview-gemma-lora-q4f16_1-MLC"  # değiştir
create_repo(HF_REPO, repo_type="model", exist_ok=True)

api = HfApi()
api.upload_folder(
    folder_path=MERGED,
    repo_id=HF_REPO,
    repo_type="model",
)
print("Uploaded:", HF_REPO)
print("Use: HF://" + HF_REPO)
```

---

## Adım 3 — Lokal PC (az disk)

`.env.hybrid`:

```env
MLC_MODEL=HF://YOUR_USERNAME/inferreview-gemma-lora-q4f16_1-MLC
# MLC_LORA_ADAPTER kullanma
```

Render:

```env
MLC_LLM_MODEL=inferreview-gemma-lora-q4f16_1-MLC
```

Stack:

```powershell
cd masterfabric-go\deployments
docker compose -f docker-compose.hybrid.yml -f docker-compose.hybrid.real-mlc.yml `
  --env-file .env.hybrid --profile tunnel up -d --force-recreate mlc-engine mlc-llm
```

İlk indirme ~2 GB (normal Gemma cache gibi). **Lokal merge yok.**

---

## Sıradaki adımlar (biz birlikte)

| # | Ne | Nerede |
|---|-----|--------|
| 1 | HF hesabı + token | huggingface.co |
| 2 | Colab notebook aç, T4 GPU | colab.research.google.com |
| 3 | `sample_train.jsonl` yükle | Bu repo |
| 4 | Hücre A→H sırayla çalıştır | Colab |
| 5 | `MLC_MODEL=HF://...` | `.env.hybrid` |
| 6 | Render model adını güncelle | Render dashboard |

Adım 1–2’yi bitirince haber ver; bir sonraki mesajda Colab hücre hücre gideriz.
