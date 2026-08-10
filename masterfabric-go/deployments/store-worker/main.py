"""Store worker: search apps and crawl Play / App Store reviews for InferReview audits."""
from __future__ import annotations

import os
import re
from datetime import datetime, timezone
from typing import Any

import httpx
from fastapi import FastAPI, HTTPException, Query
from google_play_scraper import Sort, app as play_app, reviews, search
from pydantic import BaseModel, Field

app = FastAPI(title="InferReview Store Worker", version="1.0")

DEFAULT_LANG = os.environ.get("STORE_DEFAULT_LANG", "tr")
DEFAULT_COUNTRY = os.environ.get("STORE_DEFAULT_COUNTRY", "tr")


def _resolve_play_app_id(
    query: str,
    hit: dict[str, Any],
    lang: str,
    country: str,
    cache: dict[str, str],
) -> str:
    app_id = hit.get("appId") or ""
    if app_id:
        return app_id
    slug = re.sub(r"[^a-z0-9]", "", query.lower())
    if not slug:
        return ""
    hit_title = (hit.get("title") or "").lower()
    developer = (hit.get("developer") or "").lower()
    if slug not in hit_title and slug not in developer:
        return ""
    cache_key = f"{slug}:{lang}:{country}"
    if cache_key in cache:
        return cache[cache_key]
    candidate = f"com.{slug}.app"
    try:
        info = play_app(candidate, lang=lang, country=country)
        cache[cache_key] = info.get("appId") or candidate
    except Exception:
        cache[cache_key] = ""
    return cache[cache_key]


def _iso(dt: Any) -> str | None:
    if dt is None:
        return None
    if isinstance(dt, datetime):
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        return dt.astimezone(timezone.utc).isoformat()
    return str(dt)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/search")
def search_apps(
    store: str = Query(..., pattern="^(play|appstore)$"),
    q: str = Query(..., min_length=1),
    limit: int = Query(8, ge=1, le=20),
    country: str = Query(DEFAULT_COUNTRY),
    lang: str = Query(DEFAULT_LANG),
) -> dict[str, Any]:
    q = q.strip()
    country = (country or DEFAULT_COUNTRY).strip().lower()
    lang = (lang or DEFAULT_LANG).strip().lower()
    if store == "play":
        hits = search(q, lang=lang, country=country, n_hits=min(limit * 2, 20))
        apps = []
        id_cache: dict[str, str] = {}
        for h in hits:
            app_id = _resolve_play_app_id(q, h, lang, country, id_cache)
            if not app_id:
                continue
            apps.append(
                {
                    "store": "play",
                    "app_id": app_id,
                    "app_name": h.get("title") or "",
                    "developer": h.get("developer") or "",
                    "icon_url": h.get("icon") or "",
                }
            )
            if len(apps) >= limit:
                break
        return {"store": "play", "query": q, "apps": apps}

    with httpx.Client(timeout=30.0) as client:
        resp = client.get(
            "https://itunes.apple.com/search",
            params={
                "term": q,
                "entity": "software",
                "country": country,
                "limit": limit,
            },
        )
        resp.raise_for_status()
        payload = resp.json()

    apps = []
    for item in payload.get("results", []):
        apps.append(
            {
                "store": "appstore",
                "app_id": str(item.get("trackId") or ""),
                "app_name": item.get("trackName") or "",
                "developer": item.get("artistName") or "",
                "icon_url": item.get("artworkUrl100") or "",
            }
        )
    return {"store": "appstore", "query": q, "apps": apps}


def _fetch_play(app_id: str, app_name: str, limit: int, lang: str, country: str) -> list[dict[str, Any]]:
    if not app_id:
        return []
    out: list[dict[str, Any]] = []
    token = None
    while len(out) < limit:
        batch_size = min(200, limit - len(out))
        batch, token = reviews(
            app_id,
            lang=lang,
            country=country,
            sort=Sort.NEWEST,
            count=batch_size,
            continuation_token=token,
        )
        for row in batch:
            text = (row.get("content") or "").strip()
            if not text:
                continue
            rating = int(row.get("score") or 3)
            out.append(
                {
                    "store": "play",
                    "store_review_id": f"play:{row.get('reviewId') or row.get('at') or len(out)}",
                    "app_name": app_name,
                    "rating": max(1, min(5, rating)),
                    "text": text,
                    "reviewed_at": _iso(row.get("at")),
                    "language": lang,
                }
            )
            if len(out) >= limit:
                break
        if not batch or token is None:
            break
    return out


def _fetch_appstore(app_id: str, app_name: str, limit: int, country: str) -> list[dict[str, Any]]:
    if not app_id:
        return []
    out: list[dict[str, Any]] = []
    page = 1
    with httpx.Client(timeout=30.0) as client:
        while len(out) < limit and page <= 50:
            url = (
                f"https://itunes.apple.com/{country}/rss/customerreviews/"
                f"page={page}/id={app_id}/sortby=mostrecent/json"
            )
            resp = client.get(url)
            if resp.status_code >= 400:
                break
            payload = resp.json()
            entries = payload.get("feed", {}).get("entry", [])
            if not entries:
                break
            start = 1 if page == 1 and entries and "im:rating" not in entries[0] else 0
            for entry in entries[start:]:
                rating_obj = entry.get("im:rating", {})
                rating = int(rating_obj.get("label", 0) if isinstance(rating_obj, dict) else rating_obj or 0)
                text = ""
                content = entry.get("content", {})
                if isinstance(content, dict):
                    text = (content.get("label") or "").strip()
                elif isinstance(content, str):
                    text = content.strip()
                title = entry.get("title", {})
                if not text and isinstance(title, dict):
                    text = (title.get("label") or "").strip()
                if not text:
                    continue
                review_id = entry.get("id", {})
                rid = review_id.get("label") if isinstance(review_id, dict) else str(review_id)
                updated = entry.get("updated", {})
                reviewed_at = updated.get("label") if isinstance(updated, dict) else None
                rating = max(1, min(5, rating)) if rating else 3
                out.append(
                    {
                        "store": "appstore",
                        "store_review_id": f"appstore:{rid or len(out)}",
                        "app_name": app_name,
                        "rating": rating,
                        "text": text,
                        "reviewed_at": reviewed_at,
                        "language": country,
                    }
                )
                if len(out) >= limit:
                    break
            page += 1
    return out


class CrawlRequest(BaseModel):
    app_name: str
    play_app_id: str = ""
    appstore_app_id: str = ""
    limit: int = Field(500, ge=1, le=10000)
    play_limit: int | None = None
    appstore_limit: int | None = None
    per_store_limit: int | None = None
    lang: str = DEFAULT_LANG
    country: str = DEFAULT_COUNTRY


@app.post("/crawl")
def crawl(req: CrawlRequest) -> dict[str, Any]:
    app_name = req.app_name.strip()
    if not app_name:
        raise HTTPException(status_code=400, detail="app_name required")
    if not req.play_app_id and not req.appstore_app_id:
        raise HTTPException(status_code=400, detail="at least one store app id required")

    stores = sum(1 for x in (req.play_app_id, req.appstore_app_id) if x)
    fallback_per_store = max(1, req.limit // stores) if stores else req.limit
    play_cap = req.play_limit if req.play_limit is not None else (req.per_store_limit or fallback_per_store)
    appstore_cap = req.appstore_limit if req.appstore_limit is not None else (req.per_store_limit or fallback_per_store)

    play_rows: list[dict[str, Any]] = []
    appstore_rows: list[dict[str, Any]] = []

    if req.play_app_id:
        try:
            play_rows = _fetch_play(req.play_app_id, app_name, play_cap, req.lang, req.country)
        except Exception as exc:  # noqa: BLE001
            raise HTTPException(status_code=502, detail=f"play crawl failed: {exc}") from exc

    if req.appstore_app_id:
        try:
            appstore_rows = _fetch_appstore(req.appstore_app_id, app_name, appstore_cap, req.country)
        except Exception as exc:  # noqa: BLE001
            raise HTTPException(status_code=502, detail=f"appstore crawl failed: {exc}") from exc

    combined = play_rows + appstore_rows
    combined.sort(key=lambda r: r.get("reviewed_at") or "", reverse=True)
    truncated = len(combined) > req.limit
    if truncated:
        combined = combined[: req.limit]

    return {
        "app_name": app_name,
        "play_app_id": req.play_app_id,
        "appstore_app_id": req.appstore_app_id,
        "play_count": len(play_rows),
        "appstore_count": len(appstore_rows),
        "total_returned": len(combined),
        "truncated": truncated,
        "reviews": combined,
    }
