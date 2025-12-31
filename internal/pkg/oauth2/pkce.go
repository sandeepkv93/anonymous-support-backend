package oauth2

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// generatePKCE generates a PKCE code verifier and challenge
func generatePKCE() (verifier string, challenge string, err error) {
	// Generate random bytes for code verifier (43-128 characters after encoding)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 URL encode without padding
	verifier = base64.RawURLEncoding.EncodeToString(bytes)

	// Create SHA256 hash of verifier
	hash := sha256.Sum256([]byte(verifier))

	// Base64 URL encode the hash without padding
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}
