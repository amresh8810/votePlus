// Package websocket manages WebSocket connections for real-time poll result broadcasting.
//
// Architecture (to be implemented in Phase 4):
//
//	┌──────────────┐      vote       ┌─────────────┐    pub/sub    ┌────────────┐
//	│  React Client│ ──────────────► │  Go/Gin API │ ────────────► │   Redis    │
//	│              │ ◄────────────── │  (REST)     │               │ (pub/sub)  │
//	│  WebSocket   │   live update   │             │ ◄──────────── │            │
//	└──────────────┘                 │  WebSocket  │               └────────────┘
//	                                 │  Hub        │
//	                                 └─────────────┘
//
// Key components planned:
//   - hub.go     — A central Hub goroutine that manages all active client connections,
//     subscribes to Redis channels, and fans out result messages.
//   - client.go  — Represents a single WebSocket connection; reads/writes JSON frames.
//   - handler.go — Gin HTTP upgrade handler: upgrades the connection and registers the client.
//
// Redis pub/sub ensures that in a multi-instance (horizontally scaled) deployment,
// a vote recorded on server A is broadcast to clients connected to server B.
package websocket
