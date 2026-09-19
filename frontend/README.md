# Live Polling Application — React Frontend

A production-ready, responsive single-page React application for creating live polls, sharing audience links, and observing real-time vote updates dynamically without page refreshes.

## 🛠 Technology Stack

| Library / Tool | Version | Purpose |
|----------------|---------|---------|
| **React** | 19.2 | Core UI component framework |
| **React Router** | 7.18 | Single Page Application (SPA) client-side routing |
| **Vite** | 8.3 | Lightning-fast development server & production bundler |
| **Lucide React** | 1.46 | Modern SVG icons |

---

## 🔒 Authentication & Token Architecture

```
Application Launch
       │
       ▼
POST /api/v1/auth/refresh (Browser sends HttpOnly refresh_token cookie)
       │
       ├──► [Success] ──► Receive new Access Token (Stored in runtime memory only)
       │                  Fetch GET /api/v1/auth/me to populate user profile
       │
       └──► [Failure] ──► User remains unauthenticated (Redirect to /login for protected routes)
```

> **Security Rule**: The short-lived JWT **Access Token** is kept **strictly in runtime memory** (`AuthContext` React state) and is never written to `localStorage` or `sessionStorage`. Long-lived **Refresh Tokens** are managed exclusively by the browser as secure `HttpOnly` cookies.

---

## 📡 WebSockets & Live Results Architecture

Public poll viewers and creators dynamically observe vote updates in real time:

1. On page load for `/p/:id` or `/polls/:id`, the `useWebSocket` hook opens a persistent WebSocket connection to:
   ```http
   ws://localhost:8080/api/v1/public/polls/:id/ws
   ```
2. When any user submits a vote via `POST /api/v1/public/polls/:id/vote`, the Go backend publishes the update to Redis Pub/Sub.
3. The backend WebSocket Hub broadcasts the updated JSON payload to all connected clients:
   ```json
   {
     "type": "poll_update",
     "pollId": "650000000000000000000001",
     "counts": {
       "option_1": 12,
       "option_2": 8
     },
     "totalVotes": 20
   }
   ```
4. Connected clients update options, vote counts, and progress bars instantly without reloading the page.

---

## 📁 Application Routes

| Path | Access Level | Description |
|------|--------------|-------------|
| `/login` | Public | Creator login page |
| `/signup` | Public | Creator registration page |
| `/forgot-password` | Public | Request password reset token |
| `/reset-password` | Public | Reset password with token |
| `/verify-email` | Public | One-time email verification route |
| `/p/:id` | **Public Audience** | Audience voting & live result stream (No auth required) |
| `/dashboard` | **Protected Creator** | List creator's polls, manage status, copy links |
| `/polls/create` | **Protected Creator** | Create a new poll (2-6 options) |
| `/polls/:id` | **Protected Creator** | Creator live management & analytics page |

---

## ⚙️ Environment Variables

Create `.env` in the `frontend` root directory:

```ini
# Backend API Base URL
VITE_API_URL=http://localhost:8080/api/v1
```

---

## 🚀 Getting Started

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Start Development Server
```bash
npm run dev
# App will run at http://localhost:5173
```

### 3. Build for Production
```bash
npm run build
```

---

## 🧪 Two-Browser Real-Time Voting Test

1. Ensure the Go backend server is running (`cd backend && go run ./cmd/server`).
2. Start frontend (`cd frontend && npm run dev`).
3. Log in as a creator at `http://localhost:5173/login`.
4. Create a poll at `/polls/create` with options (e.g. Python, Go, JavaScript).
5. Copy the Public Poll URL (e.g. `http://localhost:5173/p/<pollID>`).
6. Open **Browser Window A** (Normal) and **Browser Window B** (Incognito Window) at that URL.
7. Observe that both windows show `🟢 Live` connection status.
8. Submit a vote in **Browser A**.
9. Observe that **Browser B updates results instantly** without page refresh!
