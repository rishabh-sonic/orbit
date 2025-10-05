# Orbit

Go backend for a community forum, rewritten from OpenIsle (Spring Boot).

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Router | chi v5 |
| Database | PostgreSQL 16 (via pgx v5 + sqlc) |
| Cache | Redis 7 |
| Message Queue | RabbitMQ 3 |
| Search | OpenSearch 2 |
| File Storage | MinIO (S3-compatible, local) |
| Email | Resend |
| Web Push | VAPID (webpush-go) |
| Auth | JWT + OAuth (Google, GitHub, Discord, Twitter, Telegram) |

## Services

- `cmd/api` — REST API server (default port `8080`)
- `cmd/ws` — WebSocket notification delivery (default port `8082`)

## Quick Start

### 1. Copy environment config

```bash
cp .env.example .env
# edit .env as needed
```

### 2. Start infrastructure

```bash
make up
# Starts: postgres, redis, rabbitmq, opensearch, minio
```

### 3. Run migrations

```bash
# Install golang-migrate
brew install golang-migrate

make migrate
```

### 4. Run services

```bash
# In two terminals:
make api
make ws
```

Or build binaries:
```bash
make build
./bin/api
./bin/ws
```

## API Overview

All routes are under `/api`.

### Auth
| Method | Path | Auth |
|---|---|---|
| POST | `/auth/register` | — |
| POST | `/auth/verify` | — |
| POST | `/auth/login` | — |
| GET | `/auth/check` | optional |
| POST | `/auth/forgot/send` | — |
| POST | `/auth/forgot/verify` | — |
| POST | `/auth/forgot/reset` | — |
| POST | `/auth/google` | — |
| POST | `/auth/github` | — |
| POST | `/auth/discord` | — |
| POST | `/auth/twitter` | — |
| POST | `/auth/telegram` | — |

### Users
| Method | Path | Auth |
|---|---|---|
| GET | `/users/{identifier}` | — |
| GET | `/users/me` | required |
| PUT | `/users/me` | required |
| POST | `/users/me/avatar` | required |
| GET | `/users/{identifier}/following` | — |
| GET | `/users/{identifier}/followers` | — |

### Posts
| Method | Path | Auth |
|---|---|---|
| GET | `/posts` | — |
| GET | `/posts/recent` | — |
| GET | `/posts/featured` | — |
| GET | `/posts/{id}` | — |
| POST | `/posts` | required |
| PUT | `/posts/{id}` | required |
| DELETE | `/posts/{id}` | required |
| POST | `/posts/{id}/close` | required |
| POST | `/posts/{id}/reopen` | required |

### Comments
| Method | Path | Auth |
|---|---|---|
| GET | `/posts/{postId}/comments` | — |
| POST | `/posts/{postId}/comments` | required |
| GET | `/comments/{id}/replies` | — |
| POST | `/comments/{id}/replies` | required |
| DELETE | `/comments/{id}` | required |

### Notifications
| Method | Path | Auth |
|---|---|---|
| GET | `/notifications` | required |
| GET | `/notifications/unread-count` | required |
| POST | `/notifications/read` | required |
| GET/POST | `/notifications/prefs` | required |
| GET/POST | `/notifications/email-prefs` | required |

### Messages (DMs)
| Method | Path | Auth |
|---|---|---|
| GET | `/messages/conversations` | required |
| POST | `/messages` | required |
| GET | `/messages/conversations/{id}` | required |
| GET | `/messages/conversations/{id}/messages` | required |
| POST | `/messages/conversations/{id}/messages` | required |
| POST | `/messages/conversations/{id}/read` | required |
| GET | `/messages/unread-count` | required |

### Search
| Method | Path | Auth |
|---|---|---|
| GET | `/search/posts?q=...` | — |
| GET | `/search/posts/title?q=...` | — |
| GET | `/search/posts/content?q=...` | — |
| GET | `/search/users?q=...` | — |
| GET | `/search/global?q=...` | — |

### Admin (ADMIN role required)
| Method | Path |
|---|---|
| GET/POST | `/admin/config` |
| GET | `/admin/users` |
| POST | `/admin/users/{id}/ban` |
| POST | `/admin/users/{id}/unban` |
| DELETE | `/admin/posts/{id}` |
| POST | `/admin/posts/{id}/pin` |
| POST | `/admin/posts/{id}/unpin` |
| DELETE | `/admin/comments/{id}` |
| POST | `/admin/comments/{id}/pin` |
| POST | `/admin/comments/{id}/unpin` |
| GET | `/admin/stats/dau` |
| GET | `/admin/stats/dau-range?from=YYYY-MM-DD&to=YYYY-MM-DD` |
| GET | `/admin/stats/new-users-range?from=...&to=...` |
| GET | `/admin/stats/posts-range?from=...&to=...` |

## WebSocket

Connect to `ws://localhost:8082/api/ws` with `Authorization: Bearer <token>` header. The server pushes notification payloads as JSON:

```json
{
  "notification_id": "...",
  "type": "COMMENT_REPLY",
  "user_id": "...",
  "username": "...",
  "post_id": "...",
  "content": "..."
}
```

## Notification Types

| Type | Trigger |
|---|---|
| `COMMENT_REPLY` | Reply to your post or comment |
| `POST_UPDATED` | New comment on a post you subscribed to |
| `USER_FOLLOWED` | Someone followed you |
| `FOLLOWED_POST` | Someone you follow published a post |
| `USER_ACTIVITY` | User you follow created content |
| `POST_DELETED` | Admin deleted your post |
| `MENTION` | You were mentioned in a post or comment |

## File Storage

Uses MinIO locally (S3-compatible). Access the MinIO console at `http://localhost:9001` (credentials: `minioadmin` / `minioadmin`).

To switch to Tencent COS or AWS S3 in production, update these env vars:
```
STORAGE_ENDPOINT=cos.ap-guangzhou.myqcloud.com
STORAGE_ACCESS_KEY=...
STORAGE_SECRET_KEY=...
STORAGE_USE_SSL=true
STORAGE_BASE_URL=https://your-bucket.cos.ap-guangzhou.myqcloud.com
```
