package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware for structured request logging via log/slog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		if path == "/health" {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				slog.Error("request error",
					"method", c.Request.Method,
					"path", path,
					"status", status,
					"latency_ms", latency.Milliseconds(),
					"client_ip", c.ClientIP(),
					"error", e.Error(),
				)
			}
			return
		}

		slog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
