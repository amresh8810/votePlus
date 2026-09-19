// Package routes wires all HTTP and WebSocket routes to their controllers.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/controllers"
	"github.com/live-polling-app/backend/database"
	"github.com/live-polling-app/backend/middleware"
	"github.com/live-polling-app/backend/services"
	"github.com/live-polling-app/backend/websocket"
)

// Setup creates and returns the configured Gin engine.
func Setup(cfg *config.Config, db *database.DB, redisClient *database.RedisClient, wsHub *websocket.Hub) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// ── Global middleware ────────────────────────────────────────────────────
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigin: cfg.AllowedOrigin,
	}))

	// ── Initialize Services & Controllers ────────────────────────────────────
	authService := services.NewAuthService(db, cfg)
	authController := controllers.NewAuthController(authService, cfg)

	pollService := services.NewPollService(db)
	pollController := controllers.NewPollController(pollService)

	voteService := services.NewVoteService(db, redisClient, wsHub)
	voteController := controllers.NewVoteController(voteService)

	// Rate limiter for sensitive authentication endpoints
	rateLimitReqs := 20
	rateLimitWindow := cfg.RateLimitWindow
	if cfg.RateLimitAuthRequests > 0 {
		rateLimitReqs = cfg.RateLimitAuthRequests
	}
	if rateLimitWindow == 0 {
		rateLimitWindow = cfg.RateLimitWindow
	}
	authLimiter := middleware.RateLimiter(rateLimitReqs, rateLimitWindow)

	// ── Health check ─────────────────────────────────────────────────────────
	router.GET("/health", controllers.Health)

	// ── /api/v1 group ────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	{
		v1.GET("", func(c *gin.Context) {
			middleware.RespondNotFound(c)
		})

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", authLimiter, authController.Signup)
			auth.POST("/login", authLimiter, authController.Login)
			auth.POST("/refresh", authLimiter, authController.Refresh)
			auth.POST("/logout", authController.Logout)
			auth.POST("/verify-email", authController.VerifyEmail)
			auth.POST("/forgot-password", authLimiter, authController.ForgotPassword)
			auth.POST("/reset-password", authController.ResetPassword)
			auth.GET("/me", middleware.RequireAuth(cfg), authController.Me)
		}

		// Authenticated Poll Management routes
		polls := v1.Group("/polls")
		polls.Use(middleware.RequireAuth(cfg))
		{
			polls.POST("", pollController.CreatePoll)
			polls.GET("", pollController.GetMyPolls)
			polls.GET("/analytics/votes-over-time", pollController.GetVoteTimeline)
			polls.GET("/:id", pollController.GetPollByID)
			polls.PUT("/:id", pollController.UpdatePoll)
			polls.PATCH("/:id/close", pollController.ClosePoll)
			polls.DELETE("/:id", pollController.DeletePoll)
		}

		// Public Voting & WebSocket Real-Time routes
		public := v1.Group("/public/polls")
		public.Use(middleware.VoterIdentity(cfg.Env == "production")) // Attach or issue anonymous voter identity cookie
		{
			public.GET("/:id", voteController.GetPublicPoll)
			public.POST("/:id/vote", voteController.SubmitVote)
			public.GET("/:id/ws", websocket.ServeWebSocket(wsHub, voteService))
		}
	}

	router.NoRoute(func(c *gin.Context) {
		middleware.RespondNotFound(c)
	})

	return router
}
