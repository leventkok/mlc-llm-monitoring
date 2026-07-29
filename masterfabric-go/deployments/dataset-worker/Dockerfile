FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir --upgrade pip

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py push_to_hub.py ./
COPY seed/ /app/seed/

ENV HF_DATASET_ID=levonov/inferreview-app-reviews
ENV PORT=8090

EXPOSE 8090

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8090"]
