package websocket_test

import (
	"encoding/json"
	"testing"

	"github.com/live-polling-app/backend/websocket"
)

func TestPollUpdateMessageJSONFormat(t *testing.T) {
	counts := map[string]int64{
		"opt-1": 10,
		"opt-2": 5,
		"opt-3": 2,
	}
	pollID := "650000000000000000000001"

	msg := websocket.NewPollUpdateMessage(pollID, counts)

	if msg.Type != "poll_update" {
		t.Errorf("expected type 'poll_update', got '%s'", msg.Type)
	}

	if msg.PollID != pollID {
		t.Errorf("expected pollId '%s', got '%s'", pollID, msg.PollID)
	}

	if msg.TotalVotes != 17 {
		t.Errorf("expected totalVotes 17, got %d", msg.TotalVotes)
	}

	jsonBytes, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize message to JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to parse generated JSON: %v", err)
	}

	if parsed["type"] != "poll_update" {
		t.Errorf("expected type 'poll_update' in JSON, got %v", parsed["type"])
	}

	if parsed["pollId"] != pollID {
		t.Errorf("expected pollId '%s' in JSON, got %v", pollID, parsed["pollId"])
	}

	totalVotesFloat, ok := parsed["totalVotes"].(float64)
	if !ok || int64(totalVotesFloat) != 17 {
		t.Errorf("expected totalVotes 17 in JSON, got %v", parsed["totalVotes"])
	}
}

func TestHubRegistrationAndClientCount(t *testing.T) {
	hub := websocket.NewHub(nil)
	pollID := "650000000000000000000002"

	if hub.ClientCount(pollID) != 0 {
		t.Errorf("expected client count 0 for unregistered poll, got %d", hub.ClientCount(pollID))
	}
}
