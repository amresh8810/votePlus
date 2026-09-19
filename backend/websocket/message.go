// Package websocket manages real-time WebSocket client connections and Redis Pub/Sub integration.
package websocket

import (
	"encoding/json"
	"time"
)

// PollUpdateMessage represents the standard JSON frame sent to WebSocket clients.
type PollUpdateMessage struct {
	Type         string           `json:"type"`       // Always "poll_update"
	PollID       string           `json:"pollId"`     // Poll ID hex string
	Counts       map[string]int64 `json:"counts"`     // Map of option_id -> count
	TotalVotes   int64            `json:"totalVotes"` // Total votes cast
	LastActivity *time.Time       `json:"lastActivity,omitempty"`
}

// NewPollUpdateMessage constructs a PollUpdateMessage and calculates totalVotes.
func NewPollUpdateMessage(pollID string, counts map[string]int64) *PollUpdateMessage {
	var total int64
	for _, count := range counts {
		total += count
	}

	return &PollUpdateMessage{
		Type:       "poll_update",
		PollID:     pollID,
		Counts:     counts,
		TotalVotes: total,
	}

}

// NewPollUpdateMessageAt includes the persisted vote timestamp for activity views.
func NewPollUpdateMessageAt(pollID string, counts map[string]int64, lastActivity time.Time) *PollUpdateMessage {
	message := NewPollUpdateMessage(pollID, counts)
	message.LastActivity = &lastActivity
	return message
}

// ToJSON serializes the PollUpdateMessage into JSON bytes.
func (m *PollUpdateMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
