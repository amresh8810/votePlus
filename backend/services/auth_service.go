// Package services contains the core business logic, password hashing, JWT operations, and MongoDB interactions.
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/database"
	"github.com/live-polling-app/backend/models"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// SignupRequest defines input structure for POST /api/v1/auth/signup.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest defines input structure for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyEmailRequest defines input structure for POST /api/v1/auth/verify-email.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ForgotPasswordRequest defines input structure for POST /api/v1/auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest defines input structure for POST /api/v1/auth/reset-password.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// UserResponse defines the safe user representation in REST API responses.
type UserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// LoginResponse defines response structure for POST /api/v1/auth/login and refresh.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// CustomClaims represents JWT payload claims.
type CustomClaims struct {
	jwt.RegisteredClaims
}

// AuthService encapsulates authentication, password hashing, JWT, and user management logic.
type AuthService struct {
	db  *database.DB
	cfg *config.Config
}

// NewAuthService constructs a new AuthService.
func NewAuthService(db *database.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

func generateRandomHexToken(byteLen int) (string, error) {
	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// HashPassword hashes a raw password string using bcrypt with DefaultCost.
func HashPassword(password string) (string, error) {
	if len(password) > 72 {
		return "", fmt.Errorf("%w: password cannot exceed 72 bytes", ErrValidation)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a bcrypt hashed password with a raw password string.
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateToken generates a signed JWT token containing the user ID subject and expiration.
func GenerateToken(userID string, secret string, expiration time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret cannot be empty")
	}

	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and verifies a signed JWT token string using the configured HMAC secret.
func ValidateToken(tokenStr string, secret string) (*jwt.RegisteredClaims, error) {
	if secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &claims.RegisteredClaims, nil
}

// NormalizeEmail trims whitespace and converts email to lowercase.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateSignup performs strict validation on signup inputs.
func ValidateSignup(req SignupRequest) error {
	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len(trimmedName) < 2 || len(trimmedName) > 100 {
		return fmt.Errorf("%w: name must be between 2 and 100 characters", ErrValidation)
	}

	normalizedEmail := NormalizeEmail(req.Email)
	if normalizedEmail == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	if !emailRegex.MatchString(normalizedEmail) {
		return fmt.Errorf("%w: invalid email address format", ErrValidation)
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrValidation)
	}
	if len(req.Password) > 72 {
		return fmt.Errorf("%w: password cannot exceed 72 characters", ErrValidation)
	}

	return nil
}

// ValidateLogin checks basic non-empty requirements for login requests.
func ValidateLogin(req LoginRequest) error {
	normalizedEmail := NormalizeEmail(req.Email)
	if normalizedEmail == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	if req.Password == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	return nil
}

// Signup processes user registration, hashes the password, and creates the MongoDB User document.
func (s *AuthService) Signup(ctx context.Context, req SignupRequest) (*UserResponse, error) {
	if s.db == nil || s.db.UserCollection() == nil {
		return nil, ErrDBNotConnected
	}

	if err := ValidateSignup(req); err != nil {
		return nil, err
	}

	normalizedEmail := NormalizeEmail(req.Email)

	var existingUser models.User
	err := s.db.UserCollection().FindOne(ctx, bson.M{"email": normalizedEmail}).Decode(&existingUser)
	if err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("database query error: %w", err)
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		ID:            primitive.NewObjectID(),
		Name:          strings.TrimSpace(req.Name),
		Email:         normalizedEmail,
		PasswordHash:  hashedPassword,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = s.db.UserCollection().InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to save user to database: %w", err)
	}

	return &UserResponse{
		ID:            user.ID.Hex(),
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// Login verifies credentials, returns access token, raw refresh token (for cookie), and user details.
func (s *AuthService) Login(ctx context.Context, req LoginRequest, ipAddress, userAgent string) (*LoginResponse, string, error) {
	if s.db == nil || s.db.UserCollection() == nil {
		return nil, "", ErrDBNotConnected
	}

	if err := ValidateLogin(req); err != nil {
		return nil, "", err
	}

	normalizedEmail := NormalizeEmail(req.Email)

	var user models.User
	err := s.db.UserCollection().FindOne(ctx, bson.M{"email": normalizedEmail}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("database query error: %w", err)
	}

	if !CheckPassword(user.PasswordHash, req.Password) {
		return nil, "", ErrInvalidCredentials
	}

	accessExp := s.cfg.AccessTokenExpiration
	if accessExp == 0 {
		accessExp = s.cfg.JWTExpiration
	}
	accessToken, err := GenerateToken(user.ID.Hex(), s.cfg.JWTSecret, accessExp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefreshToken, _, err := s.CreateSession(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create refresh session: %w", err)
	}

	loginResp := &LoginResponse{
		Token: accessToken,
		User: UserResponse{
			ID:            user.ID.Hex(),
			Name:          user.Name,
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		},
	}

	return loginResp, rawRefreshToken, nil
}

// CreateSession creates a new hashed refresh session document in MongoDB and returns the unhashed raw token string.
func (s *AuthService) CreateSession(ctx context.Context, userID primitive.ObjectID, ipAddress, userAgent string) (string, *models.RefreshSession, error) {
	if s.db == nil || s.db.RefreshSessionCollection() == nil {
		return "", nil, ErrDBNotConnected
	}

	rawToken, err := generateRandomHexToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	tokenHash := hashToken(rawToken)
	familyID, _ := generateRandomHexToken(16)
	now := time.Now()
	refreshExp := s.cfg.RefreshTokenExpiration
	if refreshExp == 0 {
		refreshExp = 7 * 24 * time.Hour
	}

	session := models.RefreshSession{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		FamilyID:   familyID,
		TokenHash:  tokenHash,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(refreshExp),
		LastUsedAt: now,
	}

	_, err = s.db.RefreshSessionCollection().InsertOne(ctx, session)
	if err != nil {
		return "", nil, fmt.Errorf("failed to save refresh session: %w", err)
	}

	return rawToken, &session, nil
}

// RefreshTokens validates the refresh token hash, rotates the token, and returns a new access token and new raw refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context, rawRefreshToken string, ipAddress, userAgent string) (string, string, *UserResponse, error) {
	if s.db == nil || s.db.RefreshSessionCollection() == nil {
		return "", "", nil, ErrDBNotConnected
	}

	if strings.TrimSpace(rawRefreshToken) == "" {
		return "", "", nil, ErrInvalidRefreshToken
	}

	tokenHash := hashToken(rawRefreshToken)

	var session models.RefreshSession
	err := s.db.RefreshSessionCollection().FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", nil, ErrInvalidRefreshToken
		}
		return "", "", nil, fmt.Errorf("database query error: %w", err)
	}

	now := time.Now()

	// Detect token reuse / theft
	if session.RevokedAt != nil {
		slog.Warn("refresh token reuse detected! revoking entire token family", "family_id", session.FamilyID, "user_id", session.UserID.Hex())
		_, _ = s.db.RefreshSessionCollection().UpdateMany(
			ctx,
			bson.M{"family_id": session.FamilyID, "revoked_at": nil},
			bson.M{"$set": bson.M{"revoked_at": now}},
		)
		return "", "", nil, ErrInvalidRefreshToken
	}

	if session.ExpiresAt.Before(now) {
		return "", "", nil, ErrInvalidRefreshToken
	}

	// 1. Mark existing session rotated / revoked
	_, err = s.db.RefreshSessionCollection().UpdateOne(
		ctx,
		bson.M{"_id": session.ID},
		bson.M{"$set": bson.M{"revoked_at": now, "last_used_at": now}},
	)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to rotate session: %w", err)
	}

	// 2. Generate new refresh token in same family
	newRawToken, err := generateRandomHexToken(32)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newTokenHash := hashToken(newRawToken)
	refreshExp := s.cfg.RefreshTokenExpiration
	if refreshExp == 0 {
		refreshExp = 7 * 24 * time.Hour
	}

	newSession := models.RefreshSession{
		ID:         primitive.NewObjectID(),
		UserID:     session.UserID,
		FamilyID:   session.FamilyID,
		TokenHash:  newTokenHash,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(refreshExp),
		LastUsedAt: now,
	}

	_, err = s.db.RefreshSessionCollection().InsertOne(ctx, newSession)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to save new session: %w", err)
	}

	// 3. Generate short-lived access token
	accessExp := s.cfg.AccessTokenExpiration
	if accessExp == 0 {
		accessExp = 15 * time.Minute
	}
	accessToken, err := GenerateToken(session.UserID.Hex(), s.cfg.JWTSecret, accessExp)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 4. Fetch user response
	userResp, err := s.GetUserByID(ctx, session.UserID.Hex())
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, newRawToken, userResp, nil
}

// RevokeSession marks a refresh token session revoked by its raw token string.
func (s *AuthService) RevokeSession(ctx context.Context, rawRefreshToken string) error {
	if s.db == nil || s.db.RefreshSessionCollection() == nil {
		return ErrDBNotConnected
	}

	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}

	tokenHash := hashToken(rawRefreshToken)
	now := time.Now()

	_, err := s.db.RefreshSessionCollection().UpdateOne(
		ctx,
		bson.M{"token_hash": tokenHash, "revoked_at": nil},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	return err
}

// RevokeAllUserSessions revokes all active refresh sessions for a user ID hex.
func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userIDHex string) error {
	if s.db == nil || s.db.RefreshSessionCollection() == nil {
		return ErrDBNotConnected
	}

	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return ErrValidation
	}

	now := time.Now()
	_, err = s.db.RefreshSessionCollection().UpdateMany(
		ctx,
		bson.M{"user_id": objectID, "revoked_at": nil},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	return err
}

// GenerateEmailVerificationToken creates a 1-time email verification token.
func (s *AuthService) GenerateEmailVerificationToken(ctx context.Context, userIDHex string) (string, error) {
	if s.db == nil || s.db.VerificationTokenCollection() == nil {
		return "", ErrDBNotConnected
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return "", ErrValidation
	}

	rawToken, err := generateRandomHexToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	now := time.Now()

	tokenDoc := models.VerificationToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	_, err = s.db.VerificationTokenCollection().InsertOne(ctx, tokenDoc)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

// VerifyEmail verifies an email verification token and marks user email_verified=true.
func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	if s.db == nil || s.db.VerificationTokenCollection() == nil {
		return ErrDBNotConnected
	}

	if strings.TrimSpace(rawToken) == "" {
		return ErrInvalidToken
	}

	tokenHash := hashToken(rawToken)
	var vToken models.VerificationToken
	err := s.db.VerificationTokenCollection().FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&vToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrInvalidToken
		}
		return err
	}

	if vToken.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	now := time.Now()
	_, err = s.db.UserCollection().UpdateOne(
		ctx,
		bson.M{"_id": vToken.UserID},
		bson.M{"$set": bson.M{"email_verified": true, "email_verified_at": now}},
	)
	if err != nil {
		return err
	}

	_, _ = s.db.VerificationTokenCollection().DeleteOne(ctx, bson.M{"_id": vToken.ID})
	return nil
}

// ForgotPassword creates a password reset token for the given email (returns token string for dev/logging).
func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	if s.db == nil || s.db.UserCollection() == nil {
		return "", ErrDBNotConnected
	}

	normalizedEmail := NormalizeEmail(email)
	if normalizedEmail == "" {
		return "", fmt.Errorf("%w: email is required", ErrValidation)
	}
	if !emailRegex.MatchString(normalizedEmail) {
		return "", fmt.Errorf("%w: invalid email address format", ErrValidation)
	}

	var user models.User
	err := s.db.UserCollection().FindOne(ctx, bson.M{"email": normalizedEmail}).Decode(&user)
	if err != nil {
		// Return generic success to prevent account enumeration
		return "", nil
	}

	rawToken, err := generateRandomHexToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	now := time.Now()

	resetDoc := models.PasswordResetToken{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	_, err = s.db.PasswordResetCollection().InsertOne(ctx, resetDoc)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

// ResetPassword verifies password reset token, updates password hash, invalidates reset token, and revokes all active refresh sessions.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken string, newPassword string) error {
	if s.db == nil || s.db.UserCollection() == nil {
		return ErrDBNotConnected
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("%w: new password must be at least 8 characters long", ErrValidation)
	}

	tokenHash := hashToken(rawToken)
	var resetDoc models.PasswordResetToken
	err := s.db.PasswordResetCollection().FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&resetDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrInvalidToken
		}
		return err
	}

	if resetDoc.UsedAt != nil {
		return ErrTokenUsed
	}

	now := time.Now()
	if resetDoc.ExpiresAt.Before(now) {
		return ErrTokenExpired
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 1. Update user password
	_, err = s.db.UserCollection().UpdateOne(
		ctx,
		bson.M{"_id": resetDoc.UserID},
		bson.M{"$set": bson.M{"password_hash": hashedPassword, "updated_at": now}},
	)
	if err != nil {
		return err
	}

	// 2. Mark reset token used
	_, _ = s.db.PasswordResetCollection().UpdateOne(
		ctx,
		bson.M{"_id": resetDoc.ID},
		bson.M{"$set": bson.M{"used_at": now}},
	)

	// 3. Revoke all active sessions for this user
	_ = s.RevokeAllUserSessions(ctx, resetDoc.UserID.Hex())

	return nil
}

// GetUserByID retrieves a user document by ID and returns safe user information.
func (s *AuthService) GetUserByID(ctx context.Context, userIDHex string) (*UserResponse, error) {
	if s.db == nil || s.db.UserCollection() == nil {
		return nil, ErrDBNotConnected
	}

	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID format", ErrValidation)
	}

	var user models.User
	err = s.db.UserCollection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	return &UserResponse{
		ID:            user.ID.Hex(),
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}, nil
}
