package services_test

import (
	"testing"

	"github.com/live-polling-app/backend/services"
)

func TestVoterTokenDeterministicObjectIDMapping(t *testing.T) {
	tokenA := "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"
	tokenB := "b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1"

	objectIDA1 := services.VoterTokenToObjectID(tokenA)
	objectIDA2 := services.VoterTokenToObjectID(tokenA)
	objectIDB := services.VoterTokenToObjectID(tokenB)

	// Same token must yield the EXACT same ObjectID
	if objectIDA1 != objectIDA2 {
		t.Fatalf("expected deterministic ObjectID for same token, got %v and %v", objectIDA1, objectIDA2)
	}

	// Different tokens must yield different ObjectIDs
	if objectIDA1 == objectIDB {
		t.Fatalf("expected different ObjectIDs for different tokens, got same: %v", objectIDA1)
	}
}

func TestGenerateSecureVoterToken(t *testing.T) {
	token1, err := services.GenerateSecureVoterToken()
	if err != nil {
		t.Fatalf("failed to generate secure voter token: %v", err)
	}

	token2, err := services.GenerateSecureVoterToken()
	if err != nil {
		t.Fatalf("failed to generate secure voter token: %v", err)
	}

	if len(token1) != 64 {
		t.Errorf("expected 64 hex characters for 32-byte token, got length %d", len(token1))
	}

	if token1 == token2 {
		t.Errorf("expected unique random tokens, got identical tokens")
	}
}
