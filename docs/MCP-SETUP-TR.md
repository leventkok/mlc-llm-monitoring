# MCP Kurulum Rehberi (Türkçe)

Bu proje **3 MCP** kullanır: Render, Vercel, MasterFabric Academy.

Proje config: [`.cursor/mcp.json`](../.cursor/mcp.json)

---

## Adım 1 — Render MCP

### 1.1 API key al

1. [Render Dashboard → Account Settings → API Keys](https://dashboard.render.com/u/settings#api-keys)
2. **Create API Key** → kopyala (`rnd_...`)

### 1.2 Cursor'a ekle

1. Proje kökünde `.cursor/mcp.env` oluştur (örnek: [`.cursor/mcp.env.example`](../.cursor/mcp.env.example)):

```
RENDER_API_KEY=rnd_BURAYA_YAPIŞTIR
```

2. **Cursor → Settings → Tools & MCP**
3. `render` sunucusunun yeşil olduğunu kontrol et
4. Cursor'u yeniden başlat

### 1.3 İlk komut (chat'te yaz)

```
Set my Render workspace to [WORKSPACE_ADIN]
List my Render services
```

### 1.4 Bu proje için örnek prompt'lar

```
Show logs for mlc-llm-monitoring backend service
Verify ALLOWED_ORIGINS includes https://mlc-llm-monitoring.vercel.app
Redeploy the backend service
```

---

## Adım 2 — Vercel MCP

### 2.1 Otomatik kurulum (önerilen)

[Vercel MCP — Add to Cursor](https://vercel.com/docs/agent-resources/vercel-mcp) sayfasındaki **Add to Cursor** butonuna tıkla.

Veya proje `.cursor/mcp.json` içinde zaten var:

```json
"vercel": {
  "url": "https://mcp.vercel.com"
}
```

### 2.2 Login

1. **Cursor → Settings → Tools & MCP**
2. `vercel` yanında **Needs login** → tıkla
3. Vercel hesabınla authorize et

### 2.3 İlk komut

```
List my Vercel projects
Show latest production deployment for mlc-llm-monitoring
```

### 2.4 Bu proje için örnek prompt'lar

```
Show env vars for mlc-llm-monitoring project
Verify NEXT_PUBLIC_API_URL is set to my Render backend URL
Redeploy production
```

---

## Adım 3 — MasterFabric Academy MCP

### Kurulum

1. [one-hundered-days](https://github.com/masterfabric/one-hundered-days) reposunu klonla
2. `mcp` klasöründe `npm install` çalıştır
3. `~/.cursor/mcp.json` veya proje `.cursor/mcp.json` içine ekle (yolu kendi makinene göre düzenle):

```json
"masterfabric-academy": {
  "command": "C:\\path\\to\\one-hundered-days\\mcp\\node_modules\\.bin\\tsx.cmd",
  "args": ["C:\\path\\to\\one-hundered-days\\mcp\\mcp.ts"]
}
```

> Windows'ta `cwd` bazen uygulanmaz — `tsx.cmd` ve `mcp.ts` için **tam yol** kullan.

4. Cursor'u yeniden başlat

### Persona yükleme (bağlandıktan sonra)

Chat'te yaz:

```
get_mentor_persona staff-engineer
get_mentor_persona security-coach
get_academy_skill
```

Sonra:

```
staff-engineer ve security-coach personalarıyla auth endpoint'lerimi ve CORS ayarlarımı incele
```

---

## Adım 4 — Hepsini doğrula

Cursor chat'te sırayla dene:

| MCP | Test prompt |
|-----|-------------|
| Render | `List my Render services` |
| Vercel | `List my Vercel projects` |
| Academy | `get_mentor_persona staff-engineer` |

Üçü de cevap verirse MCP kurulumu tamam.

---

## MCP ile deployment workflow

```
1. Academy MCP  → staff-engineer + security-coach yükle
2. Git push     → main
3. Render MCP   → deploy status + /health kontrol
4. Vercel MCP   → NEXT_PUBLIC_API_URL + redeploy
5. Academy MCP  → security review
```

---

## Sık hatalar

| Hata | Çözüm |
|------|--------|
| Render MCP bağlanmıyor | `RENDER_API_KEY` `.cursor/mcp.env` içinde mi? Cursor restart |
| Vercel Needs login | MCP settings'ten login |
| Academy timeout | Yanlış `cwd` — path düzelt |
| Render workspace yok | `Set my Render workspace to ...` |

---

## Referanslar

- [Render MCP Docs](https://render.com/docs/mcp-server)
- [Vercel MCP Docs](https://vercel.com/docs/agent-resources/vercel-mcp)
- [Cursor MCP Docs](https://cursor.com/docs/mcp)
