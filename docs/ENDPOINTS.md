# API Endpoints (20)

## Common (1)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |

## Config (2)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/config` | No | App config |
| PUT | `/config` | No | Update config |

## Auth (9)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register (email, username, password) |
| POST | `/auth/login` | No | Login (email, password) |
| GET | `/auth/me` | Yes | Current user profile |
| PATCH | `/auth/me` | Yes | Update username |
| DELETE | `/auth/me` | Yes | Delete account + user data |
| POST | `/auth/logout` | No | Logout hint |
| POST | `/auth/refresh` | Yes | Refresh JWT |
| GET | `/auth/validate` | Yes | Validate token |
| POST | `/auth/change-password` | Yes | Change password |

## WEB MLC-LLM / Reviews (8)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/reviews` | Yes | List user's reviews |
| POST | `/reviews` | Yes | Create review |
| GET | `/reviews/{id}` | Yes | Get single review |
| GET | `/decisions` | Yes | List user's decisions |
| POST | `/decisions` | Yes | Save LLM decision |
| GET | `/scores` | Yes | List user's scores |
| POST | `/scores` | Yes | Score a decision |
| GET | `/metrics` | Yes | User metrics |

Note: LLM inference runs in the browser via `@mlc-ai/web-llm` (Gemma). Backend stores reviews, decisions, and scores.

**Total: 20 endpoints**
