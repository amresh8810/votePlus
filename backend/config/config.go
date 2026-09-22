// Package config loads and exposes application configuration.
// All values are read from environment variables, with safe defaults for local development.
// Never commit real secrets — use a .env file (git-ignored) or a secrets manager in production.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds every setting the application needs to run.
// Fields are exported so they can be read by other packages.
type Config struct {
	// Server
	Port string // HTTP port the Gin engine listens on (default: 8080)
	Env  string // "development" | "production"

	// CORS
	// AllowedOrigin is the single origin permitted to make cross-origin requests.
	// Example: "http://localhost:5173" for the Vite dev server.
	AllowedOrigin string

	// MongoDB
	MongoURI string // Full connection string, e.g. mongodb://localhost:27017
	DBName   string // Database name, e.g. "live_polling"
	SkipDB   bool   // If true, bypasses MongoDB connection during startup (useful for offline dev/testing)

	// Redis Realtime Cache & Pub/Sub
	RedisAddr     string // host:port, e.g. "localhost:6379"
	RedisPassword string // empty string means no auth
	RedisDB       int    // Redis logical database index (default: 0)
	SkipRedis     bool   // If true, bypasses Redis connection during startup (useful for offline dev/testing)

	// JWT Authentication
	JWTSecret              string        // HMAC secret for signing tokens — MUST be set in production
	JWTExpiration          time.Duration // Token lifespan, e.g. 24h
	AccessTokenExpiration  time.Duration // Short-lived access token lifespan (default: 15m)
	RefreshTokenExpiration time.Duration // Long-lived refresh session lifespan (default: 7d)

	// Rate Limiting
	RateLimitAuthRequests int           // Maximum requests per window per IP for auth routes (default: 20)
	RateLimitWindow       time.Duration // Time window for rate limiting (default: 1m)

	// Password Reset
	ExposeResetToken bool // If true, include reset tokens in responses for controlled testing only

}

// Load reads configuration from environment variables and returns a Config.
// It attempts to load a .env file from current, parent, or grandparent directory.
func Load() (*Config, error) {
	// Best-effort .env load checking current dir, parent dir (for cmd/server), and root dir
	_ = godotenv.Load(".env", "../.env", "../../.env")

	// MONGODB_URI with fallback to MONGO_URI
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	}

	// MONGODB_DATABASE with fallback to MONGO_DB_NAME
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = getEnv("MONGO_DB_NAME", "live_polling")
	}

	// REDIS_URL with fallback to REDIS_ADDR
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	}

	// REDIS_DB parsing
	redisDBInt := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsedDB, err := strconv.Atoi(dbStr); err == nil {
			redisDBInt = parsedDB
		}
	}

	// JWT Expiration parsing
	expStr := getEnv("JWT_EXPIRATION", "24h")
	jwtExpiration, err := time.ParseDuration(expStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION format '%s': %w", expStr, err)
	}

	accessExpStr := getEnv("ACCESS_TOKEN_EXPIRATION", "15m")
	accessTokenExpiration, err := time.ParseDuration(accessExpStr)
	if err != nil {
		accessTokenExpiration = 15 * time.Minute
	}

	refreshExpStr := getEnv("REFRESH_TOKEN_EXPIRATION", "168h")
	refreshTokenExpiration, err := time.ParseDuration(refreshExpStr)
	if err != nil {
		refreshTokenExpiration = 7 * 24 * time.Hour
	}

	rateLimitReqs := 20
	if reqsStr := os.Getenv("RATE_LIMIT_AUTH_REQUESTS"); reqsStr != "" {
		if parsed, err := strconv.Atoi(reqsStr); err == nil && parsed > 0 {
			rateLimitReqs = parsed
		}
	}

	rateLimitWindowStr := getEnv("RATE_LIMIT_WINDOW", "1m")
	rateLimitWindow, err := time.ParseDuration(rateLimitWindowStr)
	if err != nil {
		rateLimitWindow = 1 * time.Minute
	}

	jwtSecret := getEnv("JWT_SECRET", "dev_secret_key_change_in_production_12345")
	skipDB := os.Getenv("SKIP_DB") == "true"
	skipRedis := os.Getenv("SKIP_REDIS") == "true"
	exposeResetToken := os.Getenv("EXPOSE_RESET_TOKEN") == "true"

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		Env:                    getEnv("APP_ENV", "development"),
		AllowedOrigin:          getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		MongoURI:               mongoURI,
		DBName:                 dbName,
		SkipDB:                 skipDB,
		RedisAddr:              redisAddr,
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                redisDBInt,
		SkipRedis:              skipRedis,
		JWTSecret:              jwtSecret,
		JWTExpiration:          jwtExpiration,
		AccessTokenExpiration:  accessTokenExpiration,
		RefreshTokenExpiration: refreshTokenExpiration,
		RateLimitAuthRequests:  rateLimitReqs,
		RateLimitWindow:        rateLimitWindow,
		ExposeResetToken:       exposeResetToken,
	}

	if cfg.Env == "production" {
		if os.Getenv("JWT_SECRET") == "" || cfg.JWTSecret == "dev_secret_key_change_in_production_12345" {
			return nil, fmt.Errorf("JWT_SECRET environment variable must be explicitly set to a secure key in production")
		}
	}

	return cfg, nil
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
