# Real MLC engine — stable release wheels (0.20.0).
# Pin apache-tvm-ffi==0.1.11: 0.1.12 double-registers TVM with mlc-ai+mlc-llm (GitHub #3500).

ARG MLC_WHEEL_SUFFIX=cu130
FROM nvidia/cuda:12.8.0-devel-ubuntu22.04

ARG MLC_WHEEL_SUFFIX
ENV DEBIAN_FRONTEND=noninteractive \
    HF_HOME=/models \
    PYTHONUNBUFFERED=1 \
    CUDA_HOME=/usr/local/cuda \
    PATH=/usr/local/cuda/bin:${PATH} \
    LIBRARY_PATH=/usr/local/cuda/lib64/stubs:/usr/local/cuda/lib64 \
    TORCH_CUDA_ARCH_LIST=8.6 \
    FLASHINFER_CUDA_ARCH_LIST=8.6 \
    LD_LIBRARY_PATH=/usr/local/cuda/lib64:/usr/local/lib/python3.10/dist-packages/tvm/lib:/usr/local/lib/python3.10/dist-packages/mlc_llm/lib

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    curl \
    git \
    git-lfs \
    ca-certificates \
    build-essential \
    ninja-build \
    libx11-6 \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/* \
    && git lfs install \
    && ln -sf /usr/bin/python3 /usr/bin/python

RUN python3 -m pip install --no-cache-dir --upgrade pip \
    && python3 -m pip install --no-cache-dir \
        "apache-tvm-ffi==0.1.11" \
        pytest \
        psutil \
    && python3 -m pip install --no-cache-dir --no-deps -f https://mlc.ai/wheels \
        "mlc-ai-${MLC_WHEEL_SUFFIX}==0.20.0" \
        "mlc-llm-${MLC_WHEEL_SUFFIX}==0.20.0.dev0" \
    && python3 -m pip install --no-cache-dir \
        cloudpickle scipy tornado ml_dtypes numpy packaging typing_extensions \
        attrs datasets fastapi flashinfer-python openai pandas prompt_toolkit peft \
        requests safetensors sentencepiece shortuuid tiktoken torch tqdm transformers uvicorn aiohttp pydantic \
        huggingface_hub \
    && python3 -m pip show apache-tvm-ffi mlc-ai-${MLC_WHEEL_SUFFIX} mlc-llm-${MLC_WHEEL_SUFFIX}

EXPOSE 8081

COPY entrypoint.sh /usr/local/bin/mlc-engine-entrypoint.sh
COPY merge-lora-adapter.sh /usr/local/bin/merge-lora-adapter.sh
RUN sed -i 's/\r$//' /usr/local/bin/mlc-engine-entrypoint.sh /usr/local/bin/merge-lora-adapter.sh \
    && chmod +x /usr/local/bin/mlc-engine-entrypoint.sh /usr/local/bin/merge-lora-adapter.sh

ENTRYPOINT ["/usr/local/bin/mlc-engine-entrypoint.sh"]
