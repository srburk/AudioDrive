package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func generateToken(length int) (string, error) {
	if length <= 0 || length%2 != 0 {
		return "", fmt.Errorf("token length must be a positive even number")
	}
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func validToken(token string) bool {
	if token == "" || token == "." || token == ".." {
		return false
	}
	for _, r := range token {
		if r == '/' || r == '\\' || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func persistToken(path, token string) error {
	return os.WriteFile(path, []byte(token+"\n"), 0600)
}

// loadOrCreateToken returns the subscribe token. An explicit token is written
// to path so it survives restarts. Otherwise the existing file is reused, or
// a new token is generated and saved.
func loadOrCreateToken(path, explicit string) (string, error) {
	if explicit != "" {
		if !validToken(explicit) {
			return "", fmt.Errorf("token must be a single path segment (no slashes or whitespace)")
		}
		if err := persistToken(path, explicit); err != nil {
			return "", fmt.Errorf("write token file: %w", err)
		}
		return explicit, nil
	}

	data, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if !validToken(token) {
			return "", fmt.Errorf("token file %s is empty or invalid; replace it or pass -token", path)
		}
		return token, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("read token file: %w", err)
	}

	token, err := generateToken(TOKEN_LENGTH)
	if err != nil {
		return "", err
	}
	if err := persistToken(path, token); err != nil {
		return "", fmt.Errorf("write token file: %w", err)
	}
	return token, nil
}
