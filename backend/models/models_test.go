package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/live-polling-app/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserModelJSONSerialization(t *testing.T) {
	objectID := primitive.NewObjectID()
	now := time.Now().Truncate(time.Second)

	user := models.User{
		ID:           objectID,
		Name:         "Alice Developer",
		Email:        "alice@example.com",
		PasswordHash: "$2a$10$SecretHashedPasswordString",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal User to JSON: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Ensure password_hash is NEVER present in JSON output
	if contains(jsonStr, "SecretHashedPasswordString") || contains(jsonStr, "password_hash") {
		t.Errorf("SECURITY RISK: User JSON serialization exposed password_hash: %s", jsonStr)
	}

	// Ensure required fields are in JSON
	if !contains(jsonStr, "alice@example.com") {
		t.Errorf("expected JSON to contain email, got: %s", jsonStr)
	}
}

func TestUserModelBSONSerialization(t *testing.T) {
	objectID := primitive.NewObjectID()
	now := time.Now().Truncate(time.Second)

	user := models.User{
		ID:           objectID,
		Name:         "Bob Tester",
		Email:        "bob@example.com",
		PasswordHash: "$2a$10$HashedPasswordForMongoBSON",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Marshal to BSON
	bsonBytes, err := bson.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal User to BSON: %v", err)
	}

	// Unmarshal back from BSON
	var unmarshaledUser models.User
	if err := bson.Unmarshal(bsonBytes, &unmarshaledUser); err != nil {
		t.Fatalf("failed to unmarshal User from BSON: %v", err)
	}

	if unmarshaledUser.ID != objectID {
		t.Errorf("expected ID %v, got %v", objectID, unmarshaledUser.ID)
	}
	if unmarshaledUser.PasswordHash != "$2a$10$HashedPasswordForMongoBSON" {
		t.Errorf("expected PasswordHash to be preserved in BSON")
	}
}

func TestPollModelSerialization(t *testing.T) {
	pollID := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()

	poll := models.Poll{
		ID:       pollID,
		Question: "What is your favorite Go framework?",
		Options: []models.Option{
			{ID: "opt-1", Text: "Gin"},
			{ID: "opt-2", Text: "Echo"},
			{ID: "opt-3", Text: "Fiber"},
		},
		CreatedBy: creatorID,
		Status:    models.PollStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// JSON
	jsonBytes, err := json.Marshal(poll)
	if err != nil {
		t.Fatalf("failed to marshal Poll to JSON: %v", err)
	}

	var parsedPoll map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsedPoll); err != nil {
		t.Fatalf("failed to unmarshal JSON into map: %v", err)
	}

	if parsedPoll["question"] != "What is your favorite Go framework?" {
		t.Errorf("unexpected question in JSON: %v", parsedPoll["question"])
	}

	// BSON
	bsonBytes, err := bson.Marshal(poll)
	if err != nil {
		t.Fatalf("failed to marshal Poll to BSON: %v", err)
	}

	var bsonPoll models.Poll
	if err := bson.Unmarshal(bsonBytes, &bsonPoll); err != nil {
		t.Fatalf("failed to unmarshal BSON into Poll: %v", err)
	}

	if len(bsonPoll.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(bsonPoll.Options))
	}
	if bsonPoll.Options[0].ID != "opt-1" {
		t.Errorf("expected option ID 'opt-1', got '%s'", bsonPoll.Options[0].ID)
	}
}

func TestVoteModelSerialization(t *testing.T) {
	voteID := primitive.NewObjectID()
	pollID := primitive.NewObjectID()
	voterID := primitive.NewObjectID()

	vote := models.Vote{
		ID:        voteID,
		PollID:    pollID,
		OptionID:  "opt-1",
		VoterID:   voterID,
		CreatedAt: time.Now(),
	}

	bsonBytes, err := bson.Marshal(vote)
	if err != nil {
		t.Fatalf("failed to marshal Vote to BSON: %v", err)
	}

	var bsonVote models.Vote
	if err := bson.Unmarshal(bsonBytes, &bsonVote); err != nil {
		t.Fatalf("failed to unmarshal Vote from BSON: %v", err)
	}

	if bsonVote.OptionID != "opt-1" {
		t.Errorf("expected option_id 'opt-1', got '%s'", bsonVote.OptionID)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && stringSearch(s, substr)))
}

func stringSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
