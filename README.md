# Orbit

A full-stack community discussion platform — Go REST API + WebSocket service + Vue 3 SPA.

```
orbit/
├── backend/    Go API server (cmd/api) + WebSocket service (cmd/ws)
└── frontend/   Vue 3 + Vite + shadcn-vue SPA
```

---

## Option A — Full Docker Compose (recommended)

Builds and starts every service in one command.

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| API (direct) | http://localhost:8080 |
| WebSocket | ws://localhost:8082 |
| RabbitMQ console | http://localhost:15672 — `guest` / `guest` |
| MinIO console | http://localhost:9001 — `minioadmin` / `minioadmin` |

On first run this automatically:
1. Starts Postgres, Redis, RabbitMQ, OpenSearch and MinIO
2. Creates the MinIO `orbit` bucket
3. Runs database migrations via `golang-migrate`
4. Builds and starts the Go API + WebSocket services
5. Builds the Vue app and serves it through nginx

```bash
docker compose down      # stop, keep data volumes
docker compose down -v   # stop and wipe all data
```

---

## Option B — Infrastructure in Docker, code on host

Better for active development — infra in containers, Go and frontend run locally.

### 1. Start infrastructure

```bash
cd backend
make up        # starts postgres, redis, rabbitmq, opensearch, minio
make migrate   # requires: brew install golang-migrate
```

Create the MinIO bucket (one-time):

```bash
docker run --rm --network host minio/mc \
  sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin && \
         mc mb --ignore-existing local/orbit && \
         mc anonymous set download local/orbit"
```

### 2. Configure environment

```bash
cp .env.example backend/.env
# edit backend/.env if you want to enable optional features (see below)
```

### 3. Run the backend

```bash
cd backend
make api   # API server  → http://localhost:8080

# in a second terminal:
make ws    # WebSocket   → ws://localhost:8082

# or both at once:
make dev
```

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev   # → http://localhost:3000  (proxies /api and /oauth to localhost:8080)
```

---

## Configuration

All options have sensible defaults for local development. The table below lists the variables needed to enable optional features — copy `.env.example` to `backend/.env` and fill in only what you need.

| Feature | Environment variables |
|---|---|
| Google OAuth | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` |
| GitHub OAuth | `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` |
| Transactional email | `RESEND_API_KEY`, `RESEND_FROM_EMAIL` |
| Web push notifications | `WEBPUSH_PUBLIC_KEY`, `WEBPUSH_PRIVATE_KEY` |

Generate VAPID keys for web push:

```bash
npx web-push generate-vapid-keys
```

Register OAuth redirect URIs in each provider's developer console:

| Provider | Redirect URI |
|---|---|
| Google | `http://localhost:3000/auth/google` |
| GitHub | `http://localhost:3000/auth/github` |

---

## Tests

```bash
# Backend (no infrastructure required)
cd backend
make test

# Frontend
cd frontend
npm test
npm run test:coverage
```

---

## Stack

| Layer | Technology |
|---|---|
| API server | Go 1.25, chi router, pgx, sqlc |
| WebSocket | Go, gorilla/websocket, RabbitMQ fanout exchange |
| Auth | JWT (HS256), bcrypt, OAuth 2.0 (Google, GitHub) |
| File storage | MinIO (S3-compatible) |
| Full-text search | OpenSearch 2 |
| Frontend | Vue 3, Vite, Vue Router 4, Pinia, shadcn-vue, Tailwind CSS |
| Infrastructure | PostgreSQL 16, Redis 7, RabbitMQ 3, OpenSearch 2, MinIO |
