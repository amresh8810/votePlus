// Package models defines application data structures and BSON/JSON tags for MongoDB.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RefreshSession represents a stored, hashed refresh session in the "refresh_sessions" collection.
type RefreshSession struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	FamilyID   string             `json:"family_id" bson:"family_id"` // Tracks token chain for theft detection
	TokenHash  string             `json:"-" bson:"token_hash"`        // SHA-256 hash of the refresh token
	IPAddress  string             `json:"ip_address" bson:"ip_address"`
	UserAgent  string             `json:"user_agent" bson:"user_agent"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	ExpiresAt  time.Time          `json:"expires_at" bson:"expires_at"`
	RevokedAt  *time.Time         `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	LastUsedAt time.Time          `json:"last_used_at" bson:"last_used_at"`
}

// VerificationToken represents a one-time email verification token hash.
type VerificationToken struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	TokenHash string             `json:"-" bson:"token_hash"` // SHA-256 hash of token
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
}

// PasswordResetToken represents a one-time password reset token hash.
type PasswordResetToken struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	TokenHash string             `json:"-" bson:"token_hash"` // SHA-256 hash of token
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	UsedAt    *time.Time         `json:"used_at,omitempty" bson:"used_at,omitempty"`
}
