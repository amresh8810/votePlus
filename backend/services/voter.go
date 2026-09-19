package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateSecureVoterToken creates a cryptographically secure 32-byte hex string.
func GenerateSecureVoterToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VoterTokenToObjectID deterministically converts a voter token string into a 12-byte primitive.ObjectID.
func VoterTokenToObjectID(token string) primitive.ObjectID {
	hash := sha256.Sum256([]byte(token))
	var idBytes [12]byte
	copy(idBytes[:], hash[:12])
	return primitive.ObjectID(idBytes)
}
