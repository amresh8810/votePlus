package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/middleware"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Limit to 2 requests per 100ms for fast testing
	router.Use(middleware.RateLimiter(2, 100*time.Millisecond))
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Request 1: Should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected request 1 status 200, got %d", w1.Code)
	}

	// Request 2: Should pass
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/test", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected request 2 status 200, got %d", w2.Code)
	}

	// Request 3: Should fail with 429 Too Many Requests
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/test", nil)
	req3.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected request 3 status 429 (Too Many Requests), got %d", w3.Code)
	}
}
