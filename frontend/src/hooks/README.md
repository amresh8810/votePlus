# hooks/

Custom React hooks that encapsulate reusable stateful logic.

## Planned hooks (Phase 4+)

| File | Purpose |
|------|---------|
| `useWebSocket.js` | Manages a WebSocket connection to `/api/v1/ws/:pollID`; returns the latest results payload and connection state |
| `useAuth.js` | Reads auth context; returns `{ user, token, isAuthenticated }` |
| `usePoll.js` | Fetches and caches a single poll via the REST API |
| `usePolls.js` | Fetches the current user's list of polls |
