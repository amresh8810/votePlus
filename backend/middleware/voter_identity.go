package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/services"
)

const (
	VoterCookieName = "voter_token"
	VoterHeaderName = "X-Voter-Token"
	VoterTokenKey   = "voterToken"
)

// VoterIdentity returns a Gin middleware that extracts or issues an anonymous voter identity token.
func VoterIdentity(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. Try reading from cookie
		if cookie, err := c.Cookie(VoterCookieName); err == nil && strings.TrimSpace(cookie) != "" {
			token = strings.TrimSpace(cookie)
		}

		// 2. Fallback: Try reading from header (for non-browser HTTP clients/testing)
		if token == "" {
			if header := c.GetHeader(VoterHeaderName); strings.TrimSpace(header) != "" {
				token = strings.TrimSpace(header)
			}
		}

		// 3. If still no token, issue a new cryptographically secure voter token
		if token == "" {
			newToken, err := services.GenerateSecureVoterToken()
			if err == nil {
				token = newToken

				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(
					VoterCookieName,
					token,
					365*24*60*60, // 1 year expiry
					"/",
					"",
					isProduction, // Secure in production
					true,         // HttpOnly
				)

				c.Header(VoterHeaderName, token)
			}
		}

		if token != "" {
			c.Set(VoterTokenKey, token)
		}

		c.Next()
	}
}

// GetVoterToken retrieves the voter token from the Gin context.
func GetVoterToken(c *gin.Context) (string, bool) {
	val, exists := c.Get(VoterTokenKey)
	if !exists {
		return "", false
	}
	token, ok := val.(string)
	return token, ok
}
