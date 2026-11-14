# 🏰 Kingdom Arena

> Competitive-first, browser-based real-time tactical strategy game. Desktop browser is the first target.

## Product direction

- **Vision:** Competitive-first, 1v1 real-time tactical strategy.
- **Platform:** Web browser, desktop-first.
- **Development model:** Solo developer.
- **Board:** 12 × 10, three lanes, mirrored player territories.
- **Core roster:** 8 unit types, 3 building types and 3 resources.
- **Match duration:** 10 minutes.

## Technology

- **Frontend:** React, TypeScript, Zustand, HTML5 Canvas and Vite.
- **Backend:** Go, Chi Router and gorilla/websocket.
- **Database:** PostgreSQL 16 and Redis 7.
- **Infrastructure:** Docker Compose and Nginx.

## Run with one command (Windows)

**Requirement:** Install and start Docker Desktop. Git must be available if you are cloning the repository for the first time.

From the repository root, run:

```powershell
.\run.bat
```

The launcher builds the Go backend and React frontend images, starts PostgreSQL and Redis, waits for the web and backend health checks, and opens the game in your default browser. The default URL is `http://localhost:8088`; if that port is occupied, the launcher tries ports `8089` through `8098`.

Useful controls:

| Command | Action |
|---|---|
| `\.\run.bat` | Build, start and open the browser |
| `\.\run.bat restart` | Rebuild and restart services |
| `\.\run.bat status` | Show container status |
| `\.\run.bat logs` | Follow service logs |
| `\.\run.bat stop` | Stop containers while preserving database volumes |

The launcher uses Docker Compose. No local Go, Node.js, PostgreSQL or Redis installation is required. The first run needs an Internet connection to download base images and dependencies.

To choose a specific browser port in PowerShell:

```powershell
$env:WEB_PORT = '8090'
.\run.bat
```

## Run manually

```bash
docker compose up --build
```

The web app is served by Nginx. `/api`, `/health` and `/ws` are proxied to the Go backend.

## Phase status

- **Phase 1 — Foundation:** Browser UI, backend and infrastructure scaffold.
- **Phase 2 — Core Engine:** Server-side simulation engine, 10Hz fixed tick, game rules baseline, movement, combat, economy, buildings, bot simulation and regression tests. Merged into `main`.
- **Phase 3 — Multiplayer:** Browser-to-engine WebSocket integration, authenticated player sessions, room lifecycle, matchmaking, reconnect and live PvP remain future work.

**Important:** The current browser board is still a local prototype. It is not yet connected to the Phase 2 authoritative engine, and the current WebSocket handler is not a complete multiplayer protocol. The one-command launcher builds and opens the browser prototype plus backend services; it does not claim that live PvP is complete.

## Phase 2 engine development

Run the deterministic bot simulation:

```bash
cd backend
go run ./cmd/sim -seconds 600
```

Run backend checks:

```bash
cd backend
go test ./...
go test -race ./internal/game ./internal/room
go vet ./...
go build ./cmd/server ./cmd/sim
```

See [`docs/phase-2-rules.md`](docs/phase-2-rules.md) for the implemented initial rules baseline and explicitly deferred edge cases.
