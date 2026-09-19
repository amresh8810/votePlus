package websocket

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow cross-origin WebSocket connections from frontend
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// VoteCountProvider is an interface for getting or initializing poll vote counts.
type VoteCountProvider interface {
	GetOrInitPollVoteCounts(ctx context.Context, pollID string) (map[string]int64, error)
}

// ServeWebSocket returns a Gin handler function for GET /api/v1/public/polls/:id/ws.
func ServeWebSocket(hub *Hub, provider VoteCountProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		pollID := c.Param("id")
		if pollID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_POLL_ID", "message": "Poll ID is required in URL path"},
			})
			return
		}

		// Load or initialize current vote counts from Redis / MongoDB
		var counts map[string]int64
		if provider != nil {
			var err error
			counts, err = provider.GetOrInitPollVoteCounts(c.Request.Context(), pollID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "POLL_NOT_FOUND", "message": "Poll not found or invalid ID"},
				})
				return
			}
		} else {
			counts = make(map[string]int64)
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// Upgrade automatically responds with HTTP error if headers fail
			return
		}

		// 1. Send initial poll_update state immediately upon connection
		initialMsg := NewPollUpdateMessage(pollID, counts)
		initialJSON, err := initialMsg.ToJSON()
		if err == nil {
			_ = conn.WriteMessage(websocket.TextMessage, initialJSON)
		}

		// 2. Register client with Hub and start pumps
		client := &Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			PollID: pollID,
		}

		hub.Register(client)

		go client.writePump()
		go client.readPump()
	}
}
