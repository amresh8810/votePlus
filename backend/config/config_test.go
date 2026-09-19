package config_test

import (
	"os"
	"testing"

	"github.com/live-polling-app/backend/config"
)

func TestMongoConfigLoading(t *testing.T) {
	// Test environment variables
	os.Setenv("MONGODB_URI", "mongodb://testuser:testpass@localhost:27017/testdb")
	os.Setenv("MONGODB_DATABASE", "custom_polling_db")

	defer func() {
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("MONGODB_DATABASE")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.MongoURI != "mongodb://testuser:testpass@localhost:27017/testdb" {
		t.Errorf("expected MongoURI to match environment, got '%s'", cfg.MongoURI)
	}

	if cfg.DBName != "custom_polling_db" {
		t.Errorf("expected DBName 'custom_polling_db', got '%s'", cfg.DBName)
	}
}

func TestMongoConfigDefaultFallbacks(t *testing.T) {
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGO_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGO_DB_NAME")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("expected default MongoURI 'mongodb://localhost:27017', got '%s'", cfg.MongoURI)
	}

	if cfg.DBName != "live_polling" {
		t.Errorf("expected default DBName 'live_polling', got '%s'", cfg.DBName)
	}
}
