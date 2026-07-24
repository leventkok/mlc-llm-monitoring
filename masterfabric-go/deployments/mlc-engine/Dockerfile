# Real MLC engine — no official Docker Hub image; install from mlc.ai wheels.
# GPU passthrough via compose deploy.resources.reservations.devices (nvidia).

ARG MLC_WHEEL_SUFFIX=cu130
FROM nvidia/cuda:12.8.0-runtime-ubuntu22.04

ARG MLC_WHEEL_SUFFIX
ENV DEBIAN_FRONTEND=noninteractive \
    HF_HOME=/models \
    PYTHONUNBUFFERED=1

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    curl \
    git \
    git-lfs \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && git lfs install \
    && ln -sf /usr/bin/python3 /usr/bin/python

RUN python3 -m pip install --no-cache-dir --upgrade pip \
    && python3 -m pip install --no-cache-dir --pre -U -f https://mlc.ai/wheels \
        mlc-llm-nightly-${MLC_WHEEL_SUFFIX} \
        mlc-ai-nightly-${MLC_WHEEL_SUFFIX}

EXPOSE 8081

ENTRYPOINT ["python3", "-m", "mlc_llm", "serve"]
