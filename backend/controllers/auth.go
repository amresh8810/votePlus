// Package controllers handles HTTP requests and responses.
package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/middleware"
	"github.com/live-polling-app/backend/services"
)

const RefreshCookieName = "refresh_token"

// AuthController exposes HTTP handlers for user authentication.
type AuthController struct {
	authService *services.AuthService
	cfg         *config.Config
}

// NewAuthController constructs an AuthController.
func NewAuthController(authService *services.AuthService, cfg ...*config.Config) *AuthController {
	var c *config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return &AuthController{
		authService: authService,
		cfg:         c,
	}
}

func (ac *AuthController) setRefreshCookie(c *gin.Context, refreshToken string) {
	isProduction := false
	if ac.cfg != nil && ac.cfg.Env == "production" {
		isProduction = true
	}

	maxAge := 7 * 24 * 60 * 60 // 7 days in seconds
	if ac.cfg != nil && ac.cfg.RefreshTokenExpiration > 0 {
		maxAge = int(ac.cfg.RefreshTokenExpiration.Seconds())
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshCookieName,
		refreshToken,
		maxAge,
		"/",
		"",
		isProduction, // Secure = true only in production
		true,         // HttpOnly = true
	)
}

func (ac *AuthController) clearRefreshCookie(c *gin.Context) {
	isProduction := false
	if ac.cfg != nil && ac.cfg.Env == "production" {
		isProduction = true
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshCookieName,
		"",
		-1,
		"/",
		"",
		isProduction,
		true,
	)
}

// Signup handles POST /api/v1/auth/signup.
func (ac *AuthController) Signup(c *gin.Context) {
	var req services.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body")
		return
	}

	userResp, err := ac.authService.Signup(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		if errors.Is(err, services.ErrUserAlreadyExists) {
			middleware.RespondError(c, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "A user with this email address already exists")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	// In development, generate email verification token and log it for easy local testing
	var devVerificationToken string
	if ac.cfg == nil || ac.cfg.Env != "production" {
		if token, err := ac.authService.GenerateEmailVerificationToken(c.Request.Context(), userResp.ID); err == nil {
			devVerificationToken = token
			slog.Info("[DEV ONLY] Email verification token generated", "user_id", userResp.ID, "token", token)
		}
	}

	response := gin.H{
		"user": userResp,
	}
	if devVerificationToken != "" {
		response["dev_verification_token"] = devVerificationToken
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles POST /api/v1/auth/login.
func (ac *AuthController) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body")
		return
	}

	loginResp, rawRefreshToken, err := ac.authService.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			middleware.RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	// Set HttpOnly refresh token cookie
	if rawRefreshToken != "" {
		ac.setRefreshCookie(c, rawRefreshToken)
	}

	c.JSON(http.StatusOK, loginResp)
}

// Refresh handles POST /api/v1/auth/refresh.
func (ac *AuthController) Refresh(c *gin.Context) {
	rawRefreshToken, err := c.Cookie(RefreshCookieName)
	if err != nil || rawRefreshToken == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "MISSING_REFRESH_TOKEN", "Refresh token cookie is required")
		return
	}

	accessToken, newRawRefreshToken, userResp, err := ac.authService.RefreshTokens(c.Request.Context(), rawRefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		ac.clearRefreshCookie(c)
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			middleware.RespondError(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh session")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	// Set rotated HttpOnly refresh token cookie
	ac.setRefreshCookie(c, newRawRefreshToken)

	c.JSON(http.StatusOK, services.LoginResponse{
		Token: accessToken,
		User:  *userResp,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (ac *AuthController) Logout(c *gin.Context) {
	if rawRefreshToken, err := c.Cookie(RefreshCookieName); err == nil && rawRefreshToken != "" {
		_ = ac.authService.RevokeSession(c.Request.Context(), rawRefreshToken)
	}

	ac.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// VerifyEmail handles POST /api/v1/auth/verify-email.
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	var req services.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body")
		return
	}

	if err := ac.authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		if errors.Is(err, services.ErrInvalidToken) || errors.Is(err, services.ErrTokenExpired) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or expired verification token")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ForgotPassword handles POST /api/v1/auth/forgot-password.
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req services.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body")
		return
	}

	devToken, err := ac.authService.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	response := gin.H{
		"message": "If an account exists with that email address, password reset instructions have been generated.",
	}

	if devToken != "" && ac.cfg != nil && ac.cfg.ExposeResetToken {
		response["dev_reset_token"] = devToken
	}

	c.JSON(http.StatusOK, response)
}

// ResetPassword handles POST /api/v1/auth/reset-password.
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body")
		return
	}

	if err := ac.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, services.ErrValidation) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidToken) || errors.Is(err, services.ErrTokenExpired) || errors.Is(err, services.ErrTokenUsed) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_TOKEN", "Invalid, expired, or already used reset token")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	ac.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully. All active sessions have been revoked.",
	})
}

// Me handles GET /api/v1/auth/me (Protected route).
func (ac *AuthController) Me(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists || userID == "" {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User identity missing from request context")
		return
	}

	userResp, err := ac.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			middleware.RespondError(c, http.StatusNotFound, "USER_NOT_FOUND", "Authenticated user document not found")
			return
		}
		if errors.Is(err, services.ErrDBNotConnected) {
			middleware.RespondError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not available")
			return
		}
		middleware.RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": userResp,
	})
}
