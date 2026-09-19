// Package middleware provides HTTP middlewares including JWT authentication.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/services"
)

const (
	// ContextUserIDKey is the Gin context key where the authenticated user ID hex string is stored.
	ContextUserIDKey = "userID"
)

// RequireAuth returns a Gin middleware that validates JWT bearer tokens in the Authorization header.
// Accepts either a *config.Config or a raw string JWT secret for test convenience.
func RequireAuth(target interface{}) gin.HandlerFunc {
	var jwtSecret string

	switch t := target.(type) {
	case *config.Config:
		if t != nil {
			jwtSecret = t.JWTSecret
		}
	case string:
		jwtSecret = t
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header format must be 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token string cannot be empty")
			c.Abort()
			return
		}

		claims, err := services.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.Subject == "" {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token missing subject claim")
			c.Abort()
			return
		}

		// Store authenticated user ID in Gin context for downstream handlers
		c.Set(ContextUserIDKey, claims.Subject)
		c.Next()
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}

// GetAuthenticatedUserID is an alias for GetUserIDFromContext for compatibility.
func GetAuthenticatedUserID(c *gin.Context) (string, bool) {
	return GetUserIDFromContext(c)
}
