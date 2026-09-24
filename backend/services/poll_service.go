package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/live-polling-app/backend/database"
	"github.com/live-polling-app/backend/models"
)

var (
	ErrPollNotFound      = errors.New("poll not found")
	ErrForbidden         = errors.New("you do not have permission to access or modify this poll")
	ErrPollClosed        = errors.New("closed polls cannot be edited")
	ErrInvalidPollID     = errors.New("invalid poll ID format")
	ErrInvalidUserStatus = errors.New("invalid status query parameter")
)

type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type UpdatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type VoteTimelinePoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type OptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type PollResponse struct {
	ID        string           `json:"id"`
	Question  string           `json:"question"`
	Options   []OptionResponse `json:"options"`
	CreatedBy string           `json:"created_by"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	ViewCount int64            `json:"view_count"`
}

type PollService struct {
	db *database.DB
}

func NewPollService(db *database.DB) *PollService {
	return &PollService{db: db}
}

// ValidatePollInput validates and cleans question and options according to Phase 5 rules.
func ValidatePollInput(question string, rawOptions []string) (string, []string, error) {
	trimmedQuestion := strings.TrimSpace(question)
	qLen := len(trimmedQuestion)
	if qLen < 5 || qLen > 200 {
		return "", nil, fmt.Errorf("%w: question must be between 5 and 200 characters long", ErrValidation)
	}

	var cleanedOptions []string
	seen := make(map[string]bool)

	for _, opt := range rawOptions {
		trimmedOpt := strings.TrimSpace(opt)
		if trimmedOpt == "" {
			continue // remove empty options
		}
		lowerOpt := strings.ToLower(trimmedOpt)
		if seen[lowerOpt] {
			return "", nil, fmt.Errorf("%w: duplicate option text '%s' is not allowed", ErrValidation, trimmedOpt)
		}
		seen[lowerOpt] = true
		cleanedOptions = append(cleanedOptions, trimmedOpt)
	}

	if len(cleanedOptions) < 2 {
		return "", nil, fmt.Errorf("%w: poll must contain at least 2 non-empty options", ErrValidation)
	}

	if len(cleanedOptions) > 6 {
		return "", nil, fmt.Errorf("%w: poll cannot contain more than 6 options", ErrValidation)
	}

	return trimmedQuestion, cleanedOptions, nil
}

// FormatPollResponse converts a models.Poll document to a safe JSON PollResponse.
func FormatPollResponse(p *models.Poll) *PollResponse {
	opts := make([]OptionResponse, len(p.Options))
	for i, o := range p.Options {
		opts[i] = OptionResponse{
			ID:   o.ID,
			Text: o.Text,
		}
	}

	return &PollResponse{
		ID:        p.ID.Hex(),
		Question:  p.Question,
		Options:   opts,
		CreatedBy: p.CreatedBy.Hex(),
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		ViewCount: p.ViewCount,
	}
}

// CreatePoll handles creating a new poll for the authenticated user.
func (s *PollService) CreatePoll(ctx context.Context, userIDHex string, req CreatePollRequest) (*PollResponse, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	creatorID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid creator user ID", ErrValidation)
	}

	cleanQuestion, cleanOptions, err := ValidatePollInput(req.Question, req.Options)
	if err != nil {
		return nil, err
	}

	// Generate stable option IDs
	options := make([]models.Option, len(cleanOptions))
	for i, optText := range cleanOptions {
		options[i] = models.Option{
			ID:   fmt.Sprintf("opt-%s", primitive.NewObjectID().Hex()),
			Text: optText,
		}
	}

	now := time.Now()
	poll := models.Poll{
		ID:        primitive.NewObjectID(),
		Question:  cleanQuestion,
		Options:   options,
		CreatedBy: creatorID,
		Status:    models.PollStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = s.db.PollCollection().InsertOne(ctx, poll)
	if err != nil {
		return nil, fmt.Errorf("failed to save poll to database: %w", err)
	}

	return FormatPollResponse(&poll), nil
}

// GetMyPolls returns all polls available to authenticated users, sorted newest first.
// Optional statusFilter can be "active" or "closed".
func (s *PollService) GetMyPolls(ctx context.Context, _ string, statusFilter string) ([]*PollResponse, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	filter := bson.M{}

	if statusFilter != "" {
		normalizedStatus := strings.ToLower(strings.TrimSpace(statusFilter))
		if normalizedStatus != string(models.PollStatusActive) && normalizedStatus != string(models.PollStatusClosed) {
			return nil, fmt.Errorf("%w: status filter must be 'active' or 'closed'", ErrValidation)
		}
		filter["status"] = models.PollStatus(normalizedStatus)
	}

	// Sort newest first: created_at = -1
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.PollCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query polls: %w", err)
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, fmt.Errorf("failed to decode user polls: %w", err)
	}

	responses := make([]*PollResponse, len(polls))
	for i := range polls {
		responses[i] = FormatPollResponse(&polls[i])
	}

	return responses, nil
}

// GetVoteTimeline returns vote totals for all polls visible to the user.
func (s *PollService) GetVoteTimeline(ctx context.Context, _, period, startDate, endDate string) ([]VoteTimelinePoint, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	now := time.Now()
	end := now
	start := now
	var err error
	switch period {
	case "", "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "7d":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)
	case "30d":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)
	case "custom":
		start, err = time.ParseInLocation("2006-01-02", startDate, now.Location())
		if err != nil {
			return nil, fmt.Errorf("%w: invalid start date", ErrValidation)
		}
		endDateValue, parseErr := time.ParseInLocation("2006-01-02", endDate, now.Location())
		if parseErr != nil {
			return nil, fmt.Errorf("%w: invalid end date", ErrValidation)
		}
		end = endDateValue.AddDate(0, 0, 1).Add(-time.Nanosecond)
	default:
		return nil, fmt.Errorf("%w: period must be today, 7d, 30d, or custom", ErrValidation)
	}

	pollCursor, err := s.db.PollCollection().Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to find polls: %w", err)
	}
	defer pollCursor.Close(ctx)
	var polls []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := pollCursor.All(ctx, &polls); err != nil {
		return nil, fmt.Errorf("failed to read polls: %w", err)
	}

	hourly := period == "" || period == "today"
	points := make(map[string]int64)
	for cursor := start; !cursor.After(end); {
		if hourly {
			points[cursor.Format(time.RFC3339)] = 0
			cursor = cursor.Add(time.Hour)
		} else {
			points[cursor.Format("2006-01-02")] = 0
			cursor = cursor.AddDate(0, 0, 1)
		}
	}
	if len(polls) == 0 {
		return timelinePoints(points), nil
	}
	pollIDs := make([]primitive.ObjectID, len(polls))
	for i, poll := range polls {
		pollIDs[i] = poll.ID
	}
	voteCursor, err := s.db.VoteCollection().Find(ctx, bson.M{
		"poll_id":    bson.M{"$in": pollIDs},
		"created_at": bson.M{"$gte": start, "$lte": end},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find votes: %w", err)
	}
	defer voteCursor.Close(ctx)
	var votes []models.Vote
	if err := voteCursor.All(ctx, &votes); err != nil {
		return nil, fmt.Errorf("failed to read votes: %w", err)
	}
	for _, vote := range votes {
		voteTime := vote.CreatedAt.In(now.Location())
		key := voteTime.Format("2006-01-02")
		if hourly {
			key = voteTime.Truncate(time.Hour).Format(time.RFC3339)
		}
		if _, ok := points[key]; ok {
			points[key]++
		}
	}
	return timelinePoints(points), nil
}

func timelinePoints(points map[string]int64) []VoteTimelinePoint {
	result := make([]VoteTimelinePoint, 0, len(points))
	for date, count := range points {
		result = append(result, VoteTimelinePoint{Date: date, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result
}

// GetPollByID retrieves a poll visible to the authenticated user.
func (s *PollService) GetPollByID(ctx context.Context, pollIDHex string, _ string) (*PollResponse, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	var poll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	return FormatPollResponse(&poll), nil
}

// UpdatePoll updates question and options of an ACTIVE poll owned by the user.
func (s *PollService) UpdatePoll(ctx context.Context, pollIDHex string, userIDHex string, req UpdatePollRequest) (*PollResponse, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", ErrValidation)
	}

	cleanQuestion, cleanOptionsText, err := ValidatePollInput(req.Question, req.Options)
	if err != nil {
		return nil, err
	}

	var existingPoll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&existingPoll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	if existingPoll.CreatedBy != userID {
		return nil, ErrForbidden
	}

	// Status check: cannot edit closed poll
	if existingPoll.Status == models.PollStatusClosed {
		return nil, ErrPollClosed
	}

	// Preserve existing option IDs where text matches, generate new IDs for new options
	existingOptionMap := make(map[string]string) // text -> ID
	for _, opt := range existingPoll.Options {
		existingOptionMap[strings.ToLower(opt.Text)] = opt.ID
	}

	newOptions := make([]models.Option, len(cleanOptionsText))
	for i, optText := range cleanOptionsText {
		lowerText := strings.ToLower(optText)
		if existingID, found := existingOptionMap[lowerText]; found {
			newOptions[i] = models.Option{
				ID:   existingID,
				Text: optText,
			}
		} else {
			newOptions[i] = models.Option{
				ID:   fmt.Sprintf("opt-%s", primitive.NewObjectID().Hex()),
				Text: optText,
			}
		}
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"question":   cleanQuestion,
			"options":    newOptions,
			"updated_at": now,
		},
	}

	_, err = s.db.PollCollection().UpdateOne(ctx, bson.M{"_id": pollID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update poll in database: %w", err)
	}

	existingPoll.Question = cleanQuestion
	existingPoll.Options = newOptions
	existingPoll.UpdatedAt = now

	return FormatPollResponse(&existingPoll), nil
}

// ClosePoll marks an active poll as closed.
func (s *PollService) ClosePoll(ctx context.Context, pollIDHex string, userIDHex string) (*PollResponse, error) {
	if s.db == nil || s.db.PollCollection() == nil {
		return nil, ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return nil, ErrInvalidPollID
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", ErrValidation)
	}

	var poll models.Poll
	err = s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	if poll.CreatedBy != userID {
		return nil, ErrForbidden
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":     models.PollStatusClosed,
			"updated_at": now,
		},
	}

	_, err = s.db.PollCollection().UpdateOne(ctx, bson.M{"_id": pollID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to close poll in database: %w", err)
	}

	poll.Status = models.PollStatusClosed
	poll.UpdatedAt = now

	return FormatPollResponse(&poll), nil
}

// DeletePoll permanently removes a poll and its persisted votes.
func (s *PollService) DeletePoll(ctx context.Context, pollIDHex string, userIDHex string) error {
	if s.db == nil || s.db.PollCollection() == nil {
		return ErrDBNotConnected
	}

	pollID, err := primitive.ObjectIDFromHex(pollIDHex)
	if err != nil {
		return ErrInvalidPollID
	}
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return fmt.Errorf("%w: invalid user ID", ErrValidation)
	}
	var poll models.Poll
	if err := s.db.PollCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrPollNotFound
		}
		return fmt.Errorf("database query error: %w", err)
	}
	if poll.CreatedBy != userID {
		return ErrForbidden
	}
	if _, err := s.db.VoteCollection().DeleteMany(ctx, bson.M{"poll_id": pollID}); err != nil {
		return fmt.Errorf("failed to delete poll votes: %w", err)
	}
	if _, err := s.db.PollCollection().DeleteOne(ctx, bson.M{"_id": pollID}); err != nil {
		return fmt.Errorf("failed to delete poll: %w", err)
	}
	return nil
}
