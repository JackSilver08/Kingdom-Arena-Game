# 🏰 Kingdom Arena

> Game chiến thuật thời gian thực trên bàn cờ (Board-based RTS) kết hợp cơ chế thả quân & đẩy đường kiểu Clash Royale.

## 🚀 Công Nghệ
- **Frontend**: React 18 + TypeScript + Zustand + HTML5 Canvas (Vite)
- **Backend**: Go (Go 1.22+) + gorilla/websocket + Chi Router
- **Database**: PostgreSQL 16 + Redis 7
- **DevOps**: Docker Compose

## 🛠️ Chạy Dự Án Cục Bộ

### 1. Chạy Backend (Go)
```bash
cd backend
go run ./cmd/server
# Server lắng nghe tại http://localhost:8080 (WS: ws://localhost:8080/ws)
```

### 2. Chạy Frontend (React)
```bash
cd frontend
npm install
npm run dev
# Mở trình duyệt tại http://localhost:5173
```

### 3. Hoặc Chạy Toàn Bộ Bằng Docker
```bash
docker compose up --build
```
