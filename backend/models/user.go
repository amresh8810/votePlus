// Package models defines application data structures and BSON/JSON tags for MongoDB.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user document in the "users" collection.
type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name            string             `json:"name" bson:"name"`
	Email           string             `json:"email" bson:"email"`
	PasswordHash    string             `json:"-" bson:"password_hash"` // Never expose password hash in JSON responses
	EmailVerified   bool               `json:"email_verified" bson:"email_verified"`
	EmailVerifiedAt *time.Time         `json:"email_verified_at,omitempty" bson:"email_verified_at,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}
