"""Dataset worker: HF-backed CSV export for Go backend (no local runtime dataset)."""
from __future__ import annotations

import csv
import io
import os
from typing import Any

import httpx
from fastapi import FastAPI, Query, Response
from fastapi.responses import PlainTextResponse

app = FastAPI(title="InferReview Dataset Worker", version="1.0")

HF_SERVER = os.environ.get("HF_DATASETS_SERVER", "https://datasets-server.huggingface.co")
HF_DATASET_ID = os.environ.get("HF_DATASET_ID", "levonov/inferreview-app-reviews")
HF_TOKEN = os.environ.get("HF_TOKEN", "").strip()

COLUMNS = [
    "review_id",
    "store",
    "app_name",
    "app_version",
    "rating",
    "text",
    "language",
    "category",
    "sentiment",
]


def _headers() -> dict[str, str]:
    if HF_TOKEN:
        return {"Authorization": f"Bearer {HF_TOKEN}"}
    return {}


async def fetch_rows(length: int) -> list[dict[str, Any]]:
    url = (
        f"{HF_SERVER}/first-rows"
        f"?dataset={HF_DATASET_ID}&config=default&split=train&length={length}"
    )
    async with httpx.AsyncClient(timeout=60.0) as client:
        resp = await client.get(url, headers=_headers())
        resp.raise_for_status()
        payload = resp.json()
    rows: list[dict[str, Any]] = []
    for item in payload.get("rows", []):
        row = item.get("row") or {}
        rows.append(row)
    return rows


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok", "dataset": HF_DATASET_ID}


@app.get("/export.csv")
async def export_csv(
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=500),
) -> PlainTextResponse:
    # HF first-rows returns from start; slice after fetch for simplicity.
    need = min(offset + limit, 500)
    rows = await fetch_rows(need)
    slice_rows = rows[offset : offset + limit]

    buf = io.StringIO()
    writer = csv.DictWriter(buf, fieldnames=COLUMNS, extrasaction="ignore")
    writer.writeheader()
    for row in slice_rows:
        writer.writerow({k: row.get(k, "") for k in COLUMNS})

    return PlainTextResponse(buf.getvalue(), media_type="text/csv; charset=utf-8")


@app.get("/rows")
async def rows_json(
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=500),
) -> dict[str, Any]:
    need = min(offset + limit, 500)
    all_rows = await fetch_rows(need)
    slice_rows = all_rows[offset : offset + limit]
    return {
        "dataset": HF_DATASET_ID,
        "offset": offset,
        "limit": limit,
        "total_fetched": len(slice_rows),
        "rows": slice_rows,
    }
