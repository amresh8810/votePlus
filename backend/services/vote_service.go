package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/live-polling-app/backend/database"
	"github.com/live-polling-app/backend/models"
	"github.com/live-polling-app/backend/websocket"
)

var (
	ErrAlreadyVoted      = errors.New("you have already voted in this poll")
	ErrInvalidOption     = errors.New("the specified option does not belong to this poll")
	ErrMissingVoterToken = errors.New("voter identity token is missing")
)

type SubmitVoteRequest struct {
	OptionID string `json:"option_id"`
}

type PublicOptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type PublicPollPayload struct {
	ID        string                 `json:"id"`
	Question  string                 `json:"question"`
	Options   []PublicOptionResponse `json:"options"`
	Status    string                 `json:"status"`
	ViewCount int64                  `json:"view_count"`
}

type VoteDetails struct {
	PollID   string `json:"poll_id"`
	OptionID string `json:"option_id"`
}

type SubmitVoteResponse struct {
	Message string      `json:"message"`
	Vote    VoteDetails `json:"vote"`
}

type VoteService struct {
	db          *database.DB
	redisClient *database.RedisClient
	wsHub       *websocket.Hub
}

func NewVoteService(db *database.DB, redisClient *database.RedisClient, wsHub *websocket.Hub) *VoteService {
	return &VoteService{
		db:          db,
		redisClient: redisClient,
		wsHub:       wsHub,
	}
}

// GetPublicPoll retrieves public-facing information for a poll.
func (s *VoteService) GetPublicPoll(ctx context.Context, pollIDHex string) (*PublicPollPayload, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	if _, err := s.db.PollCollection().UpdateOne(ctx, bson.M{"_id": pollID}, bson.M{"$inc": bson.M{"view_count": 1}}); err != nil {
		return nil, fmt.Errorf("database view counter error: %w", err)
	}

	var poll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	opts := make([]PublicOptionResponse, len(poll.Options))
	for i, o := range poll.Options {
		opts[i] = PublicOptionResponse{
			ID:   o.ID,
			Text: o.Text,
		}
	}

	return &PublicPollPayload{
		ID:        poll.ID.Hex(),
		Question:  poll.Question,
		Options:   opts,
		Status:    string(poll.Status),
		ViewCount: poll.ViewCount + 1,
	}, nil
}

// GetOrInitPollVoteCounts retrieves vote counts from Redis cache. If not cached,
// aggregates vote counts from MongoDB, populates Redis cache, and returns counts.
func (s *VoteService) GetOrInitPollVoteCounts(ctx context.Context, pollIDHex string) (map[string]int64, error) {
	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	// 1. Check Redis cache if available
	if s.redisClient != nil {
		exists, err := s.redisClient.KeyExists(ctx, pollIDHex)
		if err == nil && exists {
			counts, err := s.redisClient.GetVoteCounts(ctx, pollIDHex)
			if err == nil && len(counts) > 0 {
				return counts, nil
			}
		}
	}

	// 2. Redis cache miss or uninitialized: Calculate from MongoDB
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	var poll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	counts := make(map[string]int64)
	for _, opt := range poll.Options {
		counts[opt.ID] = 0
	}

	// Count votes per option from MongoDB votes collection
	cursor, err := s.db.VoteCollection().Find(ctx, bson.M{"poll_id": pollID})
	if err == nil {
		defer cursor.Close(ctx)
		var votes []models.Vote
		if err := cursor.All(ctx, &votes); err == nil {
			for _, v := range votes {
				counts[v.OptionID]++
			}
		}
	}

	// 3. Cache calculated counts in Redis if available
	if s.redisClient != nil {
		if err := s.redisClient.SetVoteCounts(ctx, pollIDHex, counts); err != nil {
			slog.Warn("failed to cache vote counts in redis", "poll_id", pollIDHex, "error", err)
		}
	}

	return counts, nil
}

// SubmitVote validates, checks duplicates, saves vote to MongoDB (source of truth),
// increments Redis cache, and publishes a real-time Pub/Sub update.
// identityHex is the authenticated user ID; legacy voter tokens remain supported
// for service-level callers that do not have an account identity.
func (s *VoteService) SubmitVote(ctx context.Context, pollIDHex string, identityHex string, req SubmitVoteRequest) (*SubmitVoteResponse, error) {
	if s.db == nil || s.db.VoteCollection() == nil {
		return nil, ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	optionID := strings.TrimSpace(req.OptionID)
	if optionID == "" {
		return nil, fmt.Errorf("%w: option_id is required", ErrValidation)
	}

	identityHex = strings.TrimSpace(identityHex)
	if identityHex == "" {
		return nil, ErrMissingVoterToken
	}

	// 1. Fetch Poll
	var poll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 2. Check Poll Status
	if poll.Status == models.PollStatusClosed {
		return nil, ErrPollClosed
	}

	// 3. Confirm option belongs to this poll
	optionExists := false
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			optionExists = true
			break
		}
	}
	if !optionExists {
		return nil, ErrInvalidOption
	}

	// 4. Use the account ID for authenticated votes. Fall back to the
	// deterministic token mapping for legacy non-account callers.
	voterID, parseErr := primitive.ObjectIDFromHex(identityHex)
	if parseErr != nil {
		voterID = VoterTokenToObjectID(identityHex)
	}

	// 5. Application-level check for duplicate vote
	var existingVote models.Vote
	err = s.db.VoteCollection().FindOne(ctx, bson.M{
		"poll_id":  pollID,
		"voter_id": voterID,
	}).Decode(&existingVote)
	if err == nil {
		return nil, ErrAlreadyVoted
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("database query error checking existing vote: %w", err)
	}

	// 6. PERSIST VOTE IN MONGODB (Source of Truth)
	now := time.Now()
	vote := models.Vote{
		ID:        primitive.NewObjectID(),
		PollID:    pollID,
		OptionID:  optionID,
		VoterID:   voterID,
		CreatedAt: now,
	}

	_, err = s.db.VoteCollection().InsertOne(ctx, vote)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrAlreadyVoted
		}
		return nil, fmt.Errorf("failed to save vote to database: %w", err)
	}

	// 7. REAL-TIME PROCESSING (Redis HINCRBY + Pub/Sub + WebSocket Fanout)
	// Executed AFTER MongoDB success so Redis failure NEVER loses a persisted vote
	if s.redisClient != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Ensure Redis cache is initialized first if missing
			exists, _ := s.redisClient.KeyExists(bgCtx, pollIDHex)
			if !exists {
				_, _ = s.GetOrInitPollVoteCounts(bgCtx, pollIDHex)
			} else {
				_, err := s.redisClient.IncrementVoteCount(bgCtx, pollIDHex, optionID)
				if err != nil {
					slog.Error("failed to increment redis vote count", "poll_id", pollIDHex, "error", err)
				}
			}

			// Get latest counts map for real-time payload
			counts, err := s.redisClient.GetVoteCounts(bgCtx, pollIDHex)
			if err != nil || len(counts) == 0 {
				counts, _ = s.GetOrInitPollVoteCounts(bgCtx, pollIDHex)
			}

			// Create & publish JSON update
			updateMsg := websocket.NewPollUpdateMessageAt(pollIDHex, counts, now)
			jsonBytes, err := updateMsg.ToJSON()
			if err == nil {
				if err := s.redisClient.PublishUpdate(bgCtx, pollIDHex, jsonBytes); err != nil {
					slog.Error("failed to publish redis update", "poll_id", pollIDHex, "error", err)
				}
				// Also trigger direct local hub fanout for immediate client delivery
				if s.wsHub != nil {
					s.wsHub.BroadcastToPoll(pollIDHex, jsonBytes)
				}
			}
		}()
	} else if s.wsHub != nil {
		// Fallback local broadcast when Redis is skipped
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			counts, _ := s.GetOrInitPollVoteCounts(bgCtx, pollIDHex)
			updateMsg := websocket.NewPollUpdateMessageAt(pollIDHex, counts, now)
			if jsonBytes, err := updateMsg.ToJSON(); err == nil {
				s.wsHub.BroadcastToPoll(pollIDHex, jsonBytes)
			}
		}()
	}

	return &SubmitVoteResponse{
		Message: "Vote recorded successfully",
		Vote: VoteDetails{
			PollID:   pollID.Hex(),
			OptionID: optionID,
		},
	}, nil
}
