package websocket

import (
	"context"
	"log/slog"
	"sync"

	"github.com/live-polling-app/backend/database"
)

// Hub maintains the set of active WebSocket clients grouped by poll ID
// and coordinates broadcasting messages received from Redis Pub/Sub channels.
type Hub struct {
	mu          sync.RWMutex
	pollClients map[string]map[*Client]bool
	subscribers map[string]context.CancelFunc // Track active Redis Pub/Sub goroutines per poll
	redisClient *database.RedisClient
}

// NewHub constructs a new thread-safe WebSocket Hub.
func NewHub(redisClient *database.RedisClient) *Hub {
	return &Hub{
		pollClients: make(map[string]map[*Client]bool),
		subscribers: make(map[string]context.CancelFunc),
		redisClient: redisClient,
	}
}

// Register adds a new client to the hub under its specified pollID.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.pollClients[client.PollID]; !exists {
		h.pollClients[client.PollID] = make(map[*Client]bool)
	}
	h.pollClients[client.PollID][client] = true
	slog.Debug("client registered for poll updates", "poll_id", client.PollID)

	// Ensure Redis Pub/Sub subscriber goroutine is active for this pollID
	if h.redisClient != nil && h.subscribers[client.PollID] == nil {
		ctx, cancel := context.WithCancel(context.Background())
		h.subscribers[client.PollID] = cancel
		go h.listenRedisPubSub(ctx, client.PollID)
	}
}

// Unregister removes a client from the hub and cleans up empty poll maps.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.pollClients[client.PollID]; exists {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			slog.Debug("client unregistered from poll updates", "poll_id", client.PollID)
		}

		// If no clients remain for this pollID, cancel the Redis Pub/Sub subscriber goroutine to free memory
		if len(clients) == 0 {
			delete(h.pollClients, client.PollID)
			if cancel, active := h.subscribers[client.PollID]; active {
				cancel()
				delete(h.subscribers, client.PollID)
				slog.Debug("closed redis pubsub subscriber for idle poll", "poll_id", client.PollID)
			}
		}
	}
}

// BroadcastToPoll sends a JSON payload to all active WebSocket clients watching the specified pollID.
// Non-blocking: if a client's send buffer is full, it is unregistered to prevent blocking other clients.
func (h *Hub) BroadcastToPoll(pollID string, message []byte) {
	h.mu.RLock()
	clientsMap, exists := h.pollClients[pollID]
	if !exists || len(clientsMap) == 0 {
		h.mu.RUnlock()
		return
	}

	// Copy client pointers to avoid holding read lock during channel sends
	clients := make([]*Client, 0, len(clientsMap))
	for client := range clientsMap {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// Buffer full, unregister slow/broken client
			h.Unregister(client)
		}
	}
}

// ClientCount returns the number of active WebSocket clients watching a given pollID.
func (h *Hub) ClientCount(pollID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, exists := h.pollClients[pollID]; exists {
		return len(clients)
	}
	return 0
}

// listenRedisPubSub subscribes to the Redis channel poll:{pollID}:updates and broadcasts updates to clients.
func (h *Hub) listenRedisPubSub(ctx context.Context, pollID string) {
	if h.redisClient == nil {
		return
	}

	pubsub := h.redisClient.SubscribeUpdates(ctx, pollID)
	if pubsub == nil {
		return
	}
	defer func() {
		_ = pubsub.Close()
	}()

	ch := pubsub.Channel()
	slog.Info("started redis pubsub listener for poll", "poll_id", pollID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping redis pubsub listener for poll", "poll_id", pollID)
			return

		case msg, ok := <-ch:
			if !ok {
				return
			}
			if msg != nil && msg.Payload != "" {
				h.BroadcastToPoll(pollID, []byte(msg.Payload))
			}
		}
	}
}
