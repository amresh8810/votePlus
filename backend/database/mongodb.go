// Package database manages MongoDB connection lifecycle, client instantiation, and index creation.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/live-polling-app/backend/config"
)

// Collection names constants
const (
	UsersCollection              = "users"
	PollsCollection              = "polls"
	VotesCollection              = "votes"
	RefreshSessionsCollection    = "refresh_sessions"
	VerificationTokensCollection = "verification_tokens"
	PasswordResetsCollection     = "password_resets"
)

// DB wraps the MongoDB client and target database.
type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect initializes the MongoDB client, verifies connection with a ping,
// and ensures required collection indexes are created.
func Connect(ctx context.Context, cfg *config.Config) (*DB, error) {
	// Allow Atlas connection and index setup enough time on the first startup.
	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	clientOptions := mongoClientOptions(cfg.MongoURI)

	client, err := mongo.Connect(pingCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	// Ping database server to confirm connectivity
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(pingCtx)
		return nil, fmt.Errorf("failed to ping mongodb server: %w", err)
	}

	db := &DB{
		Client:   client,
		Database: client.Database(cfg.DBName),
	}

	slog.Info("connected to mongodb successfully", "database", cfg.DBName)

	// Create required database indexes
	if err := db.ensureIndexes(pingCtx); err != nil {
		_ = client.Disconnect(pingCtx)
		return nil, fmt.Errorf("failed to create database indexes: %w", err)
	}

	return db, nil
}

func mongoClientOptions(uri string) *options.ClientOptions {
	return options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(30 * time.Second).
		SetConnectTimeout(15 * time.Second).
		SetSocketTimeout(30 * time.Second)
}

// Close gracefully disconnects the MongoDB client.
func (db *DB) Close(ctx context.Context) error {
	if db.Client == nil {
		return nil
	}
	slog.Info("disconnecting from mongodb...")
	return db.Client.Disconnect(ctx)
}

// UserCollection returns the "users" mongo.Collection handle.
func (db *DB) UserCollection() *mongo.Collection {
	return db.Database.Collection(UsersCollection)
}

// PollCollection returns the "polls" mongo.Collection handle.
func (db *DB) PollCollection() *mongo.Collection {
	return db.Database.Collection(PollsCollection)
}

// VoteCollection returns the "votes" mongo.Collection handle.
func (db *DB) VoteCollection() *mongo.Collection {
	return db.Database.Collection(VotesCollection)
}

// RefreshSessionCollection returns the "refresh_sessions" mongo.Collection handle.
func (db *DB) RefreshSessionCollection() *mongo.Collection {
	return db.Database.Collection(RefreshSessionsCollection)
}

// VerificationTokenCollection returns the "verification_tokens" mongo.Collection handle.
func (db *DB) VerificationTokenCollection() *mongo.Collection {
	return db.Database.Collection(VerificationTokensCollection)
}

// PasswordResetCollection returns the "password_resets" mongo.Collection handle.
func (db *DB) PasswordResetCollection() *mongo.Collection {
	return db.Database.Collection(PasswordResetsCollection)
}

// ensureIndexes sets up all required collection indexes in MongoDB.
func (db *DB) ensureIndexes(ctx context.Context) error {
	// 1. Users collection: unique email index
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_users_email_unique"),
		},
	}
	if _, err := db.UserCollection().Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("error creating user indexes: %w", err)
	}

	// 2. Polls collection: creator lookup index
	pollIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "created_by", Value: 1}},
			Options: options.Index().SetName("idx_polls_created_by"),
		},
	}
	if _, err := db.PollCollection().Indexes().CreateMany(ctx, pollIndexes); err != nil {
		return fmt.Errorf("error creating poll indexes: %w", err)
	}

	// 3. Votes collection: poll lookup index & compound unique (poll_id + voter_id)
	voteIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "poll_id", Value: 1}},
			Options: options.Index().SetName("idx_votes_poll_id"),
		},
		{
			Keys: bson.D{
				{Key: "poll_id", Value: 1},
				{Key: "voter_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_votes_poll_voter_unique"),
		},
	}
	if _, err := db.VoteCollection().Indexes().CreateMany(ctx, voteIndexes); err != nil {
		return fmt.Errorf("error creating vote indexes: %w", err)
	}

	// 4. Refresh sessions collection: unique token_hash, user_id, family_id
	sessionIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_refresh_sessions_token_hash_unique"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_refresh_sessions_user_id"),
		},
		{
			Keys:    bson.D{{Key: "family_id", Value: 1}},
			Options: options.Index().SetName("idx_refresh_sessions_family_id"),
		},
	}
	if _, err := db.RefreshSessionCollection().Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("error creating refresh session indexes: %w", err)
	}

	// 5. Verification tokens collection: unique token_hash, user_id
	verificationIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_verification_tokens_hash_unique"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_verification_tokens_user_id"),
		},
	}
	if _, err := db.VerificationTokenCollection().Indexes().CreateMany(ctx, verificationIndexes); err != nil {
		return fmt.Errorf("error creating verification token indexes: %w", err)
	}

	// 6. Password resets collection: unique token_hash, user_id
	passwordResetIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_password_resets_hash_unique"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_password_resets_user_id"),
		},
	}
	if _, err := db.PasswordResetCollection().Indexes().CreateMany(ctx, passwordResetIndexes); err != nil {
		return fmt.Errorf("error creating password reset indexes: %w", err)
	}

	slog.Info("mongodb indexes verified successfully")
	return nil
}
