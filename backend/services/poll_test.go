package services_test

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/live-polling-app/backend/models"
	"github.com/live-polling-app/backend/services"
)

func TestValidatePollInputSuccess(t *testing.T) {
	question := "  What is your favorite programming language?  "
	rawOptions := []string{"  Go ", " Python", " JavaScript  ", ""}

	q, opts, err := services.ValidatePollInput(question, rawOptions)
	if err != nil {
		t.Fatalf("expected valid poll input to pass, got: %v", err)
	}

	if q != "What is your favorite programming language?" {
		t.Errorf("expected trimmed question, got '%s'", q)
	}

	if len(opts) != 3 {
		t.Fatalf("expected 3 non-empty cleaned options, got %d", len(opts))
	}

	if opts[0] != "Go" || opts[1] != "Python" || opts[2] != "JavaScript" {
		t.Errorf("unexpected options slice: %v", opts)
	}
}

func TestValidatePollInputShortQuestion(t *testing.T) {
	_, _, err := services.ValidatePollInput("Hi?", []string{"Option A", "Option B"})
	if err == nil {
		t.Errorf("expected validation error for question shorter than 5 characters")
	}
}

func TestValidatePollInputLongQuestion(t *testing.T) {
	longQ := string(make([]byte, 201))
	for i := range longQ {
		longQ = longQ[:i] + "a" + longQ[i+1:]
	}
	_, _, err := services.ValidatePollInput(longQ, []string{"Option A", "Option B"})
	if err == nil {
		t.Errorf("expected validation error for question longer than 200 characters")
	}
}

func TestValidatePollInputLessThanTwoOptions(t *testing.T) {
	_, _, err := services.ValidatePollInput("Valid Question?", []string{"Only One Option"})
	if err == nil {
		t.Errorf("expected validation error for less than 2 options")
	}

	// 2 raw options, but 1 is empty string after trimming
	_, _, err = services.ValidatePollInput("Valid Question?", []string{"Only One Option", "   "})
	if err == nil {
		t.Errorf("expected validation error when non-empty options count < 2")
	}
}

func TestValidatePollInputMoreThanSixOptions(t *testing.T) {
	opts := []string{"Opt 1", "Opt 2", "Opt 3", "Opt 4", "Opt 5", "Opt 6", "Opt 7"}
	_, _, err := services.ValidatePollInput("Valid Question?", opts)
	if err == nil {
		t.Errorf("expected validation error for more than 6 options")
	}
}

func TestValidatePollInputDuplicateOptions(t *testing.T) {
	// Case-insensitive duplicate check
	opts := []string{"Go", "Python", "  go  "}
	_, _, err := services.ValidatePollInput("Valid Question?", opts)
	if err == nil {
		t.Errorf("expected validation error for duplicate options (case insensitive)")
	}
}

func TestFormatPollResponse(t *testing.T) {
	pollID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	now := time.Now()

	poll := &models.Poll{
		ID:        pollID,
		Question:  "Best Go Framework?",
		Options:   []models.Option{{ID: "opt-1", Text: "Gin"}, {ID: "opt-2", Text: "Fiber"}},
		CreatedBy: userID,
		Status:    models.PollStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := services.FormatPollResponse(poll)

	if resp.ID != pollID.Hex() {
		t.Errorf("expected ID '%s', got '%s'", pollID.Hex(), resp.ID)
	}
	if resp.CreatedBy != userID.Hex() {
		t.Errorf("expected CreatedBy '%s', got '%s'", userID.Hex(), resp.CreatedBy)
	}
	if resp.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", resp.Status)
	}
	if len(resp.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(resp.Options))
	}
}
