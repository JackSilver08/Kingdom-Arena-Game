# 🏰 Kingdom Arena

> Game chiến thuật thời gian thực trên bàn cờ (Board-based RTS), kết hợp triển khai quân, kiểm soát ba lane và giao tranh tự động.

## Product direction

- **Vision:** Competitive-first, 1v1 real-time tactical strategy.
- **Platform:** Web browser, desktop-first.
- **Development model:** Solo developer.
- **Board:** 12 × 10, three lanes, mirrored player territories.
- **Core roster:** 8 unit types, 3 building types and 3 resources.
- **Match duration:** 10 minutes.

## Technology

- **Frontend:** React 18, TypeScript, Zustand, HTML5 Canvas and Vite.
- **Backend:** Go, Chi Router and gorilla/websocket.
- **Database:** PostgreSQL 16 and Redis 7.
- **Infrastructure:** Docker Compose and Nginx.

## Phase status

- **Phase 1 — Foundation:** Initial browser, backend and infrastructure scaffold.
- **Phase 2 — Core Engine:** Server-side simulation engine, game rules baseline, movement, combat, economy, buildings, bot simulation and regression tests are being delivered on the `phase-2/core-engine` branch.
- **Phase 3 — Multiplayer:** WebSocket command dispatch, authenticated player sessions, matchmaking, reconnect and live PvP integration remain future work.

The Phase 2 engine is a backend domain implementation. The existing browser prototype is not yet wired to it as a server-authoritative live client; that integration is part of the next networking milestone.

## Run locally

### Backend HTTP server

```bash
cd backend
go run ./cmd/server
```

The HTTP server listens on port `8080` by default. The current WebSocket handler is still a foundation and does not yet expose the Phase 2 engine as a complete multiplayer protocol.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

### Docker Compose

```bash
docker compose up --build
```

## Phase 2 engine development

Run the deterministic bot simulation:

```bash
cd backend
go run ./cmd/sim -seconds 600
```

Run focused engine and room tests:

```bash
go test ./internal/game ./internal/room ./cmd/sim ./pkg/models
```

Run race detection and static analysis:

```bash
go test -race ./internal/game ./internal/game ./internal/room
go vet ./internal/game ./internal/room ./cmd/sim ./pkg/models
```

The repository's GitHub Actions workflow additionally runs the complete backend test, vet and build commands using the Go version declared in `backend/go.mod`.

## Phase 2 design notes
See [`docs/phase-2-rules.md`](docs/phase-2-rules.md) for the implemented initial rules baseline and explicitly deferred edge cases.
