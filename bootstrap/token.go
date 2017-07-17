package bootstrap

import (
"crypto/rand"
"fmt"
"encoding/hex"
"github.com/kris-nova/kubicorn/logger"
"os"
)

const (
	TokenIDBytes     = 3
	TokenSecretBytes = 8
)

func GetToken() string {
	t, err := generateToken()
	if err != nil {
		logger.Critical("Unable to generate bootstrap token to authenticate nodes to master. Failing!")
		os.Exit(1)
	}
	return t
}

func randBytes(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken generates a new token with a token ID that is valid as a
// Kubernetes DNS label.
// For more info, see kubernetes/pkg/util/validation/validation.go.
func generateToken() (string, error) {
	tokenID, err := randBytes(TokenIDBytes)
	if err != nil {
		return "", err
	}

	tokenSecret, err := randBytes(TokenSecretBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", tokenID, tokenSecret), nil
}