package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/middleware"
)

func setupVoterTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.VoterIdentity(false))
	router.GET("/test-voter", func(c *gin.Context) {
		token, ok := middleware.GetVoterToken(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "voter token missing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"voter_token": token})
	})
	return router
}

func TestVoterIdentityIssuesNewCookieWhenMissing(t *testing.T) {
	router := setupVoterTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/test-voter", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	// Verify Cookie header was set
	cookies := w.Result().Cookies()
	var voterCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == middleware.VoterCookieName {
			voterCookie = c
			break
		}
	}

	if voterCookie == nil {
		t.Fatalf("expected voter_token cookie to be set")
	}

	if !voterCookie.HttpOnly {
		t.Errorf("expected voter_token cookie to be HttpOnly")
	}

	if len(voterCookie.Value) != 64 {
		t.Errorf("expected 64 hex char token, got length %d", len(voterCookie.Value))
	}
}

func TestVoterIdentityReusesExistingCookie(t *testing.T) {
	router := setupVoterTestRouter()

	existingToken := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
	req, _ := http.NewRequest(http.MethodGet, "/test-voter", nil)
	req.AddCookie(&http.Cookie{Name: middleware.VoterCookieName, Value: existingToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	if !containsStr(w.Body.String(), existingToken) {
		t.Errorf("expected response to reuse existing token '%s', got '%s'", existingToken, w.Body.String())
	}
}

func TestVoterIdentityReusesHeader(t *testing.T) {
	router := setupVoterTestRouter()

	headerToken := "ffeeddccbbaa00998877665544332211ffeeddccbbaa00998877665544332211"
	req, _ := http.NewRequest(http.MethodGet, "/test-voter", nil)
	req.Header.Set(middleware.VoterHeaderName, headerToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	if !containsStr(w.Body.String(), headerToken) {
		t.Errorf("expected response to reuse header token '%s', got '%s'", headerToken, w.Body.String())
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && stringIndex(s, substr) >= 0))
}

func stringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
