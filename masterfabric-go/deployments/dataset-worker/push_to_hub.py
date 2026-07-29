#!/usr/bin/env python3
"""Push seed CSV to Hugging Face Hub (MasterFabric Academy data pipeline)."""
from __future__ import annotations

import os
import sys
from pathlib import Path

import pandas as pd
from datasets import Dataset
from huggingface_hub import get_token


def default_seed_path() -> Path:
    docker = Path("/app/seed/reviews.csv")
    if docker.is_file():
        return docker
    return Path(__file__).resolve().parent / "seed" / "reviews.csv"


def main() -> int:
    raw = os.environ.get("SEED_CSV", "").strip()
    csv_path = Path(raw).expanduser() if raw else default_seed_path()
    repo_id = os.environ.get("HF_DATASET_ID", "levonov/inferreview-app-reviews")
    token = os.environ.get("HF_TOKEN", "").strip() or (get_token() or "")
    private = os.environ.get("HF_DATASET_PRIVATE", "false").lower() == "true"

    if not csv_path.is_file():
        print(f"seed CSV not found: {csv_path}", file=sys.stderr)
        print(f"expected at: {default_seed_path()}", file=sys.stderr)
        return 1
    if not token:
        print("HF_TOKEN is required (env var or: huggingface-cli login)", file=sys.stderr)
        return 1

    df = pd.read_csv(csv_path)
    required = {
        "review_id",
        "store",
        "app_name",
        "app_version",
        "rating",
        "text",
        "language",
        "category",
        "sentiment",
    }
    missing = required - set(df.columns)
    if missing:
        print(f"missing columns: {sorted(missing)}", file=sys.stderr)
        return 1

    ds = Dataset.from_pandas(df, preserve_index=False)
    print(f"Pushing {len(df)} rows to {repo_id}...")
    ds.push_to_hub(repo_id, token=token, private=private)
    print("Done.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
