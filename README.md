# Orbit

A full-stack forum — Go 1.25 REST API + WebSocket service + Vue 3 frontend.

```
orbit/
├── backend/    Go API server (cmd/api) + WebSocket service (cmd/ws)
└── frontend/   Vue 3 + Vite + shadcn-vue SPA
```

---

## Option A — Full Docker Compose (recommended for first run)

Builds and starts every service in containers.

```bash
# From the project root:
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend (nginx SPA) | http://localhost:3000 |
| API (direct access) | http://localhost:8080 |
| WebSocket | ws://localhost:8082 |
| RabbitMQ management | http://localhost:15672 (guest / guest) |
| MinIO console | http://localhost:9001 (minioadmin / minioadmin) |

The first run automatically:
1. Starts Postgres, Redis, RabbitMQ, OpenSearch, MinIO
2. Creates the MinIO `orbit` bucket
3. Runs DB migrations via `golang-migrate`
4. Builds and starts the Go API + WS services
5. Builds the Vue app and serves it via nginx

To stop everything:
```bash
docker compose down          # keep volumes (data persisted)
docker compose down -v       # also wipe all data volumes
```

---

## Option B — Infrastructure in Docker, services on host

Faster iteration: infra runs in containers, Go and frontend run directly on your machine.

### 1. Start infrastructure

```bash
cd backend
make up        # docker compose up -d (postgres, redis, rabbitmq, opensearch, minio)
sleep 5
make migrate   # requires: brew install golang-migrate
```

Create the MinIO bucket (one time):
```bash
docker run --rm --network host minio/mc \
  sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin && \
         mc mb --ignore-existing local/orbit && \
         mc anonymous set download local/orbit"
```

### 2. Configure environment

```bash
cd backend
cp ../.env.example .env     # edit JWT_SECRET etc. if desired
```

### 3. Run API + WebSocket servers

```bash
cd backend
make api   # → http://localhost:8080
# in another terminal:
make ws    # → ws://localhost:8082

# or both at once:
make dev
```

### 4. Run the frontend dev server

```bash
cd frontend
npm install
npm run dev    # → http://localhost:3000  (proxies /api → localhost:8080)
```

---

## Configuration

Copy `.env.example` to `backend/.env` and fill in any real keys you need:

| Feature | Variables needed |
|---|---|
| Email verification | `RESEND_API_KEY`, `RESEND_FROM_EMAIL` |
| Web Push | `WEBPUSH_PUBLIC_KEY`, `WEBPUSH_PRIVATE_KEY` |
| Google OAuth | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` |
| GitHub OAuth | `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` |
| Discord OAuth | `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET` |
| Twitter OAuth | `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET` |
| Telegram OAuth | `TELEGRAM_BOT_TOKEN` |

Everything else has working defaults for local development.

To generate VAPID keys for Web Push:
```bash
npx web-push generate-vapid-keys
```

---

## Running tests

```bash
# Backend (no infrastructure needed)
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
| API server | Go, chi router, pgx / sqlc |
| WebSocket | Go, gorilla/websocket, RabbitMQ fanout |
| Auth | JWT, bcrypt, OAuth2 (Google, GitHub, Discord, Twitter, Telegram) |
| Storage | MinIO (S3-compatible, runs locally) |
| Search | OpenSearch 2 |
| Frontend | Vue 3, Vite, Vue Router 4, Pinia, shadcn-vue, Tailwind CSS |
| Infrastructure | PostgreSQL 16, Redis 7, RabbitMQ 3, OpenSearch 2, MinIO |
