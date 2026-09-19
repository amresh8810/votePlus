package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/middleware"
	"github.com/live-polling-app/backend/services"
)

func setupAuthTestRouter(jwtSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/protected")
	protected.Use(middleware.RequireAuth(jwtSecret))
	protected.GET("", func(c *gin.Context) {
		userID, ok := middleware.GetAuthenticatedUserID(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user ID missing in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "user_id": userID})
	})

	return router
}

func TestRequireAuthMissingHeader(t *testing.T) {
	secret := "test_jwt_secret_key_32_bytes_long"
	router := setupAuthTestRouter(secret)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for missing header, got %d", w.Code)
	}
}

func TestRequireAuthMalformedHeader(t *testing.T) {
	secret := "test_jwt_secret_key_32_bytes_long"
	router := setupAuthTestRouter(secret)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic some_base64_string")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for malformed header, got %d", w.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	secret := "test_jwt_secret_key_32_bytes_long"
	router := setupAuthTestRouter(secret)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token.string")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for invalid token, got %d", w.Code)
	}
}

func TestRequireAuthExpiredToken(t *testing.T) {
	secret := "test_jwt_secret_key_32_bytes_long"
	router := setupAuthTestRouter(secret)

	expiredToken, err := services.GenerateToken("650000000000000000000001", secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for expired token, got %d", w.Code)
	}
}

func TestRequireAuthValidToken(t *testing.T) {
	secret := "test_jwt_secret_key_32_bytes_long"
	router := setupAuthTestRouter(secret)

	validUserID := "650000000000000000000099"
	validToken, err := services.GenerateToken(validUserID, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for valid token, got %d", w.Code)
	}
}
