package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/routes"
	"github.com/live-polling-app/backend/services"
)

func setupTestRouter() (http.Handler, *config.Config) {
	cfg := &config.Config{
		Port:          "8080",
		Env:           "development",
		AllowedOrigin: "http://localhost:5173",
		JWTSecret:     "unit_test_jwt_secret_key_32bytes_long",
		JWTExpiration: 1 * time.Hour,
	}
	return routes.Setup(cfg, nil, nil, nil), cfg
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
	if resp["service"] != "live-polling-backend" {
		t.Errorf("expected service 'live-polling-backend', got '%s'", resp["service"])
	}
}

func TestApiV1RootReturns404(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d for /api/v1 root, got %d", http.StatusNotFound, w.Code)
	}

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}

	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", resp.Error.Code)
	}
}

func TestUndefinedRouteReturnsStructured404(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/undefined-endpoint", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}

	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", resp.Error.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Allowed origin test
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin header to be 'http://localhost:5173', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}

	// 2. Preflight OPTIONS test
	optionsReq, _ := http.NewRequest(http.MethodOptions, "/api/v1/polls", nil)
	optionsReq.Header.Set("Origin", "http://localhost:5173")
	optionsW := httptest.NewRecorder()
	router.ServeHTTP(optionsW, optionsReq)

	if optionsW.Code != http.StatusNoContent {
		t.Errorf("expected OPTIONS preflight status 204, got %d", optionsW.Code)
	}
}

func TestProtectedMeRouteWithoutAuthHeaderReturns401(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProtectedMeRouteWithMalformedAuthHeaderReturns401(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "InvalidHeaderFormatWithoutSpace")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for malformed header, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProtectedMeRouteWithInvalidJWTSecretReturns401(t *testing.T) {
	router, _ := setupTestRouter()

	// Generate token signed with wrong secret
	tokenStr, _ := services.GenerateToken("650000000000000000000001", "wrong_secret_key_32bytes_long", 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for invalid secret token, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProtectedMeRouteWithExpiredJWTReturns401(t *testing.T) {
	router, cfg := setupTestRouter()

	// Generate token expired 1 hour ago
	tokenStr, _ := services.GenerateToken("650000000000000000000001", cfg.JWTSecret, -1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for expired token, got %d", http.StatusUnauthorized, w.Code)
	}
}
