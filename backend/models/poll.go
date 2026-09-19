package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollStatus defines valid states for a poll.
type PollStatus string

const (
	PollStatusActive PollStatus = "active"
	PollStatusClosed PollStatus = "closed"
)

// Option represents a single selectable choice in a poll.
type Option struct {
	ID   string `json:"id" bson:"id"` // Stable unique ID (e.g., UUID or short hex string)
	Text string `json:"text" bson:"text"`
}

// Poll represents a poll document stored in the "polls" collection.
type Poll struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Question  string             `json:"question" bson:"question"`
	Options   []Option           `json:"options" bson:"options"`
	CreatedBy primitive.ObjectID `json:"created_by" bson:"created_by"`
	Status    PollStatus         `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	ViewCount int64              `json:"view_count" bson:"view_count"`
}
