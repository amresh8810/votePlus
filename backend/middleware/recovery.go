package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics and logs them.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"error", rec,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errorBody{
					Error: apiError{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "An unexpected error occurred. Please try again later.",
					},
				})
			}
		}()
		c.Next()
	}
}
