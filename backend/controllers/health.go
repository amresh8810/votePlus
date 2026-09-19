// Package controllers contains the HTTP handler functions for each API resource.
// Each handler validates input, delegates to a service if needed, and writes JSON.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// healthResponse is the payload returned by GET /health.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Health handles GET /health.
// It returns HTTP 200 with status and service name.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "live-polling-backend",
	})
}
