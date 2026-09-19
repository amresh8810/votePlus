package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig specifies CORS policy settings.
type CORSConfig struct {
	AllowedOrigin string
}

// CORS returns a Gin middleware for Cross-Origin Resource Sharing.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == cfg.AllowedOrigin {
			c.Header("Access-Control-Allow-Origin", cfg.AllowedOrigin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Voter-Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
