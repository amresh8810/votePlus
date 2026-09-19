package database

import (
	"testing"
	"time"

	"github.com/live-polling-app/backend/config"
)

func TestRedisKeyFormatters(t *testing.T) {
	pollID := "650000000000000000000099"

	expectedVotesKey := "poll:650000000000000000000099:votes"
	if key := VotesKey(pollID); key != expectedVotesKey {
		t.Errorf("expected votes key '%s', got '%s'", expectedVotesKey, key)
	}

	expectedChannelKey := "poll:650000000000000000000099:updates"
	if channel := UpdatesChannel(pollID); channel != expectedChannelKey {
		t.Errorf("expected updates channel '%s', got '%s'", expectedChannelKey, channel)
	}
}

func TestRedisOptionsParsesURL(t *testing.T) {
	cfg := &config.Config{
		RedisAddr: "rediss://redis-user:redis-password@example.com:6380/4",
		RedisDB:   7,
	}

	options, err := redisOptions(cfg)
	if err != nil {
		t.Fatalf("expected Redis URL to parse: %v", err)
	}
	if options.Addr != "example.com:6380" {
		t.Fatalf("expected host and port to become example.com:6380, got %q", options.Addr)
	}
	if options.Username != "redis-user" {
		t.Fatalf("expected username to parse, got %q", options.Username)
	}
	if options.Password != "redis-password" {
		t.Fatalf("expected password to parse")
	}
	if options.DB != 7 {
		t.Fatalf("expected configured DB 7, got %d", options.DB)
	}
}

func TestRedisOptionsSupportsHostPort(t *testing.T) {
	options, err := redisOptions(&config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "local-password",
		RedisDB:       2,
	})
	if err != nil {
		t.Fatalf("expected host:port configuration to parse: %v", err)
	}
	if options.Addr != "localhost:6379" || options.Password != "local-password" || options.DB != 2 {
		t.Fatalf("host:port configuration was not preserved")
	}
}

func TestRedisOptionsRejectsInvalidURL(t *testing.T) {
	_, err := redisOptions(&config.Config{RedisAddr: "redis://[invalid"})
	if err == nil {
		t.Fatal("expected invalid Redis URL to return an error")
	}
	if got := err.Error(); len(got) < len("invalid REDIS_URL:") || got[:len("invalid REDIS_URL:")] != "invalid REDIS_URL:" {
		t.Fatalf("expected clear REDIS_URL error, got %q", got)
	}
}

func TestRedisOptionsEnablesTLSForRedissURL(t *testing.T) {
	options, err := redisOptions(&config.Config{
		RedisAddr: "rediss://example.com:6380",
	})
	if err != nil {
		t.Fatalf("expected rediss URL to parse: %v", err)
	}
	if options.TLSConfig == nil {
		t.Fatal("expected rediss URL to enable TLS")
	}
}

func TestMongoClientOptionsUseDriverManagedTLS(t *testing.T) {
	clientOptions := mongoClientOptions("mongodb+srv://user:password@example.mongodb.net/?retryWrites=true")

	if clientOptions.TLSConfig != nil {
		t.Fatal("expected mongodb+srv options not to inject a custom TLS config")
	}
	if clientOptions.ServerSelectionTimeout == nil || *clientOptions.ServerSelectionTimeout != 30*time.Second {
		t.Fatalf("expected 30 second server selection timeout, got %v", clientOptions.ServerSelectionTimeout)
	}
	if clientOptions.ConnectTimeout == nil || *clientOptions.ConnectTimeout != 15*time.Second {
		t.Fatalf("expected 15 second connect timeout, got %v", clientOptions.ConnectTimeout)
	}
	if clientOptions.SocketTimeout == nil || *clientOptions.SocketTimeout != 30*time.Second {
		t.Fatalf("expected 30 second socket timeout, got %v", clientOptions.SocketTimeout)
	}
}
