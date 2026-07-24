# Inferreview App MCP Server

Local **stdio MCP server** so Cursor agents can call the inferreview REST API (reviews, analyze, stats) without changing the live backend.

This is separate from the deployment MCPs (Render, Vercel, Academy) documented in [MCP.md](./MCP.md).

## Tools

| Tool | API | Auth |
|------|-----|------|
| `login` | `POST /auth/login` | No (establishes session) |
| `get_health` | `GET /health` | No |
| `get_stats` | `GET /stats` | Yes |
| `list_reviews` | `GET /reviews` | Yes |
| `create_review` | `POST /reviews` | Yes |
| `analyze_review` | `POST /reviews/{id}/analyze` | Yes |
| `list_decisions` | `GET /decisions` | Yes |
| `list_scores` | `GET /scores` | Yes |

All tool results are pretty-printed JSON text.

## Build

```bash
cd masterfabric-go
go build -o inferreview-mcp ./cmd/mcp
```

Or run without installing:

```bash
cd masterfabric-go
go run ./cmd/mcp
```

## Cursor setup

1. Copy [`.cursor/mcp.env.example`](../.cursor/mcp.env.example) → `.cursor/mcp.env`
2. Set credentials (pick one auth mode):

```env
INFERREVIEW_API_URL=https://mlc-llm-monitoring.onrender.com
INFERREVIEW_EMAIL=you@example.com
INFERREVIEW_PASSWORD=your-password
```

Or use a JWT bearer token instead of email/password:

```env
INFERREVIEW_JWT_TOKEN=eyJ...
```

3. Add to `.cursor/mcp.json` (merge with existing servers):

On Windows, Cursor may ignore `cwd` and `go -C`. Use the launcher script instead:

```json
"inferreview": {
  "command": "${workspaceFolder}/.cursor/run-inferreview-mcp.cmd",
  "args": [],
  "env": {
    "INFERREVIEW_API_URL": "https://mlc-llm-monitoring.onrender.com",
    "INFERREVIEW_EMAIL": "${env:INFERREVIEW_EMAIL}",
    "INFERREVIEW_PASSWORD": "${env:INFERREVIEW_PASSWORD}"
  }
}
```

Linux/macOS alternative:

```json
"args": ["-C", "masterfabric-go", "run", "./cmd/mcp"]
```

Copy-paste starter: [examples/cursor-mcp-inferreview.json](../examples/cursor-mcp-inferreview.json)

4. Restart Cursor → **Settings → Tools & MCP** → `inferreview` should be green.

## Example prompts

```
get_stats
list_reviews limit 10
create_review app_name MyApp store play rating 2 text "App crashes on startup"
analyze_review review_id <uuid-from-create>
```

## Privacy

- Credentials stay in `.cursor/mcp.env` (gitignored) or your shell env.
- MCP talks to your configured API URL only; no extra telemetry.
- Live prod is unchanged — this is a thin HTTP client over existing REST routes.
