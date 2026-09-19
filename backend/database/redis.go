// Package database manages Redis client pool, key formatting, Pub/Sub channels, and vote count cache operations.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/live-polling-app/backend/config"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	Client *redis.Client
}

func redisOptions(cfg *config.Config) (*redis.Options, error) {
	var options *redis.Options
	if strings.HasPrefix(cfg.RedisAddr, "redis://") || strings.HasPrefix(cfg.RedisAddr, "rediss://") {
		parsed, err := redis.ParseURL(cfg.RedisAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
		}
		options = parsed
	} else {
		options = &redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
		}
	}

	options.DB = cfg.RedisDB
	options.PoolSize = 10
	options.MinIdleConns = 3
	options.DialTimeout = 3 * time.Second
	options.ReadTimeout = 3 * time.Second
	options.WriteTimeout = 3 * time.Second
	return options, nil
}

// ConnectRedis initializes the Redis client, verifies connection with a ping,
// and returns a RedisClient handle.
func ConnectRedis(ctx context.Context, cfg *config.Config) (*RedisClient, error) {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	options, err := redisOptions(cfg)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to ping redis at %s: %w", options.Addr, err)
	}

	slog.Info("connected to redis successfully", "addr", options.Addr, "db", options.DB)
	return &RedisClient{Client: client}, nil
}

// Close gracefully shuts down the Redis connection pool.
func (rc *RedisClient) Close() error {
	if rc == nil || rc.Client == nil {
		return nil
	}
	slog.Info("disconnecting from redis...")
	return rc.Client.Close()
}

// VotesKey returns the Redis Hash key for a poll's vote counts: poll:{pollID}:votes
func VotesKey(pollID string) string {
	return fmt.Sprintf("poll:%s:votes", pollID)
}

// UpdatesChannel returns the Redis Pub/Sub channel key: poll:{pollID}:updates
func UpdatesChannel(pollID string) string {
	return fmt.Sprintf("poll:%s:updates", pollID)
}

// KeyExists checks if the Redis vote counts hash exists for the given pollID.
func (rc *RedisClient) KeyExists(ctx context.Context, pollID string) (bool, error) {
	if rc == nil || rc.Client == nil {
		return false, fmt.Errorf("redis client is nil")
	}
	val, err := rc.Client.Exists(ctx, VotesKey(pollID)).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// GetVoteCounts retrieves all option vote counts for a poll from Redis.
func (rc *RedisClient) GetVoteCounts(ctx context.Context, pollID string) (map[string]int64, error) {
	if rc == nil || rc.Client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	result, err := rc.Client.HGetAll(ctx, VotesKey(pollID)).Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(result))
	for optionID, strVal := range result {
		count, _ := strconv.ParseInt(strVal, 10, 64)
		counts[optionID] = count
	}
	return counts, nil
}

// SetVoteCounts sets the entire vote count map for a poll in Redis.
func (rc *RedisClient) SetVoteCounts(ctx context.Context, pollID string, counts map[string]int64) error {
	if rc == nil || rc.Client == nil {
		return fmt.Errorf("redis client is nil")
	}
	if len(counts) == 0 {
		return nil
	}

	fields := make(map[string]interface{}, len(counts))
	for optID, count := range counts {
		fields[optID] = count
	}

	key := VotesKey(pollID)
	pipe := rc.Client.Pipeline()
	pipe.HSet(ctx, key, fields)
	// Set 24 hour TTL on vote cache so inactive polls expire naturally
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// IncrementVoteCount atomically increments the vote count for an option in Redis using HINCRBY.
func (rc *RedisClient) IncrementVoteCount(ctx context.Context, pollID string, optionID string) (int64, error) {
	if rc == nil || rc.Client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}
	key := VotesKey(pollID)
	newCount, err := rc.Client.HIncrBy(ctx, key, optionID, 1).Result()
	if err != nil {
		return 0, err
	}
	// Refresh TTL on vote activity
	_ = rc.Client.Expire(ctx, key, 24*time.Hour).Err()
	return newCount, nil
}

// DeleteVoteCounts removes the cached vote counts for a poll.
func (rc *RedisClient) DeleteVoteCounts(ctx context.Context, pollID string) error {
	if rc == nil || rc.Client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return rc.Client.Del(ctx, VotesKey(pollID)).Err()
}

// PublishUpdate publishes a JSON message payload to the Redis Pub/Sub channel for a poll.
func (rc *RedisClient) PublishUpdate(ctx context.Context, pollID string, payload []byte) error {
	if rc == nil || rc.Client == nil {
		return fmt.Errorf("redis client is nil")
	}
	channel := UpdatesChannel(pollID)
	return rc.Client.Publish(ctx, channel, payload).Err()
}

// SubscribeUpdates subscribes to the Redis Pub/Sub channel for a poll.
func (rc *RedisClient) SubscribeUpdates(ctx context.Context, pollID string) *redis.PubSub {
	if rc == nil || rc.Client == nil {
		return nil
	}
	channel := UpdatesChannel(pollID)
	return rc.Client.Subscribe(ctx, channel)
}
