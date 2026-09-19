package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/middleware"
	"github.com/live-polling-app/backend/services"
)

type VoteController struct {
	voteService *services.VoteService
}

func NewVoteController(voteService *services.VoteService) *VoteController {
	return &VoteController{voteService: voteService}
}

// GetPublicPoll handles GET /api/v1/public/polls/:id.
func (vc *VoteController) GetPublicPoll(c *gin.Context) {
	pollID := c.Param("id")

	poll, err := vc.voteService.GetPublicPoll(c.Request.Context(), pollID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPollID) {
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
			return
		}
		if errors.Is(err, services.ErrPollNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

// SubmitVote handles POST /api/v1/public/polls/:id/vote.
func (vc *VoteController) SubmitVote(c *gin.Context) {
	pollID := c.Param("id")

	var req services.SubmitVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request JSON payload")
		return
	}

	voterToken, ok := middleware.GetVoterToken(c)
	if !ok || voterToken == "" {
		middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Voter identity token missing")
		return
	}

	resp, err := vc.voteService.SubmitVote(c.Request.Context(), pollID, voterToken, req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPollID) {
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
			return
		}
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidOption) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_OPTION", "The specified option does not belong to this poll.")
			return
		}
		if errors.Is(err, services.ErrPollClosed) {
			middleware.RespondError(c, http.StatusBadRequest, "POLL_CLOSED", "Voting is closed for this poll.")
			return
		}
		if errors.Is(err, services.ErrPollNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
			return
		}
		if errors.Is(err, services.ErrAlreadyVoted) {
			middleware.RespondError(c, http.StatusConflict, "ALREADY_VOTED", "You have already voted in this poll.")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
