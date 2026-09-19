package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/middleware"
	"github.com/live-polling-app/backend/services"
)

type PollController struct {
	pollService *services.PollService
}

func NewPollController(pollService *services.PollService) *PollController {
	return &PollController{pollService: pollService}
}

// CreatePoll handles POST /api/v1/polls.
func (pc *PollController) CreatePoll(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	var req services.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request JSON payload")
		return
	}

	poll, err := pc.pollService.CreatePoll(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"poll": poll})
}

// GetMyPolls handles GET /api/v1/polls?status=active|closed.
func (pc *PollController) GetMyPolls(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	statusFilter := c.Query("status")

	polls, err := pc.pollService.GetMyPolls(c.Request.Context(), userID, statusFilter)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

// GetVoteTimeline handles GET /api/v1/polls/analytics/votes-over-time.
func (pc *PollController) GetVoteTimeline(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}
	points, err := pc.pollService.GetVoteTimeline(
		c.Request.Context(),
		userID,
		c.DefaultQuery("period", "today"),
		c.Query("start"),
		c.Query("end"),
	)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		middleware.RespondInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

// GetPollByID handles GET /api/v1/polls/:id.
func (pc *PollController) GetPollByID(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	pollID := c.Param("id")

	poll, err := pc.pollService.GetPollByID(c.Request.Context(), pollID, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPollID) {
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
			return
		}
		if errors.Is(err, services.ErrPollNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			middleware.RespondError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view this poll")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

// UpdatePoll handles PUT /api/v1/polls/:id.
func (pc *PollController) UpdatePoll(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	pollID := c.Param("id")

	var req services.UpdatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request JSON payload")
		return
	}

	poll, err := pc.pollService.UpdatePoll(c.Request.Context(), pollID, userID, req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPollID) {
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
			return
		}
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, services.ErrPollClosed) {
			middleware.RespondError(c, http.StatusBadRequest, "POLL_CLOSED", "Closed polls cannot be edited")
			return
		}
		if errors.Is(err, services.ErrPollNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			middleware.RespondError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to update this poll")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

// ClosePoll handles PATCH /api/v1/polls/:id/close.
func (pc *PollController) ClosePoll(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	pollID := c.Param("id")

	poll, err := pc.pollService.ClosePoll(c.Request.Context(), pollID, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPollID) {
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
			return
		}
		if errors.Is(err, services.ErrPollNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			middleware.RespondError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to close this poll")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

// DeletePoll handles DELETE /api/v1/polls/:id.
func (pc *PollController) DeletePoll(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity not found in context")
		return
	}

	err := pc.pollService.DeletePoll(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidPollID):
			middleware.RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid poll ID format")
		case errors.Is(err, services.ErrPollNotFound):
			middleware.RespondError(c, http.StatusNotFound, "NOT_FOUND", "The requested poll was not found")
		case errors.Is(err, services.ErrForbidden):
			middleware.RespondError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to delete this poll")
		case errors.Is(err, services.ErrDBNotConnected):
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
		default:
			middleware.RespondInternalError(c)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
