# 24/7 Production: VPS + Docker + Cloudflare

PC kapalıyken de **MLC + Grafana + Prometheus** ayakta kalsın istiyorsan stack'i bir **VPS**'e taşı.

Render (API) ve Vercel (frontend) cloud'ta kalır; sadece inference + monitoring VPS'te 7/24 çalışır.

## Mimari

```
app.inferreview.com (Vercel) ──► api.inferreview.com (Render)
                                        │
                                        │ POST /reviews/{id}/analyze
                                        ▼
                              mlc.inferreview.com (Cloudflare Tunnel)
                                        │
                                        ▼
                              VPS Docker network "inferreview-prod"
                              ┌─────────────────────────────────────┐
                              │ cloudflared                         │
                              │   ├─► mlc-gateway:80 ─► mlc-llm     │
                              │   └─► grafana:3000                  │
                              │ grafana ──► prometheus:9090         │
                              │ prometheus ──► mlc-llm + Render API │
                              └─────────────────────────────────────┘
```

## Docker içi haberleşme

| From | To | URL |
|------|-----|-----|
| cloudflared | MLC | `http://mlc-gateway:80` |
| cloudflared | Grafana | `http://grafana:3000` |
| mlc-gateway | MLC | `http://mlc-llm:8080` |
| Grafana | Prometheus | `http://prometheus:9090` (provisioned) |
| Prometheus | MLC mock | `mlc-llm:8080/metrics` |
| Prometheus | Render API | `https://mlc-llm-monitoring.onrender.com/metrics` |

Hepsi aynı Docker network (`inferreview-prod`); dışarıya sadece Cloudflare Tunnel açılır.

## 1 — VPS kirala

Öneri: Hetzner CX22, DigitalOcean 2GB, vb. (~€4–6/ay)

- OS: **Ubuntu 24.04**
- Docker + Docker Compose v2 kurulu olsun

## 2 — Repo + env

```bash
git clone https://github.com/leventkok/mlc-llm-monitoring.git
cd mlc-llm-monitoring/masterfabric-go/deployments
cp .env.prod.example .env.prod
nano .env.prod   # MLC_API_KEY, CLOUDFLARE_TUNNEL_TOKEN, GRAFANA_ADMIN_PASSWORD
```

## 3 — PC'deki hybrid stack'i durdur

Aynı tunnel token iki yerde aynı anda sorun çıkarabilir:

```powershell
# Windows PC
docker compose -f docker-compose.hybrid.yml --env-file .env.hybrid down
```

## 4 — VPS'te başlat

```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod up --build -d
docker compose -f docker-compose.prod.yml ps
```

Tüm servisler `restart: unless-stopped` — VPS reboot sonrası Docker açılınca otomatik kalkar.

## 5 — Boot'ta otomatik (systemd)

```bash
sudo cp systemd/inferreview.service /etc/systemd/system/
sudo sed -i "s|/opt/inferreview/mlc-llm-monitoring|$(pwd)/../..|g" /etc/systemd/system/inferreview.service
# WorkingDirectory'yi deployments klasörüne ayarla:
sudo sed -i 's|WorkingDirectory=.*|WorkingDirectory='"$(pwd)"'|' /etc/systemd/system/inferreview.service
sudo systemctl daemon-reload
sudo systemctl enable inferreview
sudo systemctl start inferreview
```

## 6 — Render (değişmez)

| Key | Value |
|-----|-------|
| `MLC_LLM_BASE_URL` | `https://mlc.inferreview.com` |
| `MLC_LLM_API_KEY` | `.env.prod` içindeki `MLC_API_KEY` |
| `METRICS_ENABLED` | `true` |

## 7 — Test

| URL | Beklenen |
|-----|----------|
| https://mlc.inferreview.com/health | `{"status":"ok"}` |
| https://grafana.inferreview.com | Grafana login |
| Vercel → Analyze | Karar döner |
| Grafana KPI dashboard | Metrikler artar |

## 8 — Frontend + API domain (opsiyonel)

Cloudflare DNS:

| CNAME | Target |
|-------|--------|
| `app` | Vercel |
| `api` | Render custom domain |

Render: `ALLOWED_ORIGINS=https://app.inferreview.com`  
Vercel: custom domain `app.inferreview.com`

## PC vs VPS

| | PC (hybrid) | VPS (prod) |
|--|-------------|------------|
| 7/24 | ❌ PC açık olmalı | ✅ |
| Compose | `docker-compose.hybrid.yml --profile tunnel` | `docker-compose.prod.yml` |
| Tunnel | Aynı token | Aynı token |
