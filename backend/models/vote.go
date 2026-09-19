package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vote represents a recorded vote document in the "votes" collection.
// A unique index on (poll_id, voter_id) enforces one vote per authenticated user per poll.
type Vote struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PollID    primitive.ObjectID `json:"poll_id" bson:"poll_id"`
	OptionID  string             `json:"option_id" bson:"option_id"`
	VoterID   primitive.ObjectID `json:"voter_id" bson:"voter_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
