// Package middleware provides HTTP middlewares including rate limiting.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipHistory struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// RateLimiter creates an in-memory sliding window rate limiter middleware per client IP.
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	clients := make(map[string]*ipHistory)
	var clientsMu sync.Mutex

	// Periodic cleanup of stale IP records every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			now := time.Now()
			clientsMu.Lock()
			for ip, history := range clients {
				history.mu.Lock()
				validIdx := 0
				for i, ts := range history.timestamps {
					if now.Sub(ts) <= window {
						validIdx = i
						break
					}
				}
				if validIdx > 0 {
					history.timestamps = history.timestamps[validIdx:]
				}
				if len(history.timestamps) == 0 {
					delete(clients, ip)
				}
				history.mu.Unlock()
			}
			clientsMu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}

		clientsMu.Lock()
		history, exists := clients[ip]
		if !exists {
			history = &ipHistory{timestamps: make([]time.Time, 0, limit)}
			clients[ip] = history
		}
		clientsMu.Unlock()

		now := time.Now()
		history.mu.Lock()
		defer history.mu.Unlock()

		// Filter timestamps within current window
		cutoff := now.Add(-window)
		validCount := 0
		for _, ts := range history.timestamps {
			if ts.After(cutoff) {
				history.timestamps[validCount] = ts
				validCount++
			}
		}
		history.timestamps = history.timestamps[:validCount]

		if len(history.timestamps) >= limit {
			RespondError(c, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		history.timestamps = append(history.timestamps, now)
		c.Next()
	}
}
