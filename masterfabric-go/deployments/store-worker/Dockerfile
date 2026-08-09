FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir --upgrade pip

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .

ENV PORT=8091
ENV STORE_DEFAULT_LANG=tr
ENV STORE_DEFAULT_COUNTRY=tr

EXPOSE 8091

CMD ["sh", "-c", "uvicorn main:app --host 0.0.0.0 --port ${PORT:-8091}"]
