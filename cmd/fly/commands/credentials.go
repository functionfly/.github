package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials stores the authenticated user's identity and token.
type Credentials struct {
	Version      string    `json:"version"`
	User         UserInfo  `json:"user"`
	Token        string    `json:"token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserInfo holds the authenticated user's profile.
type UserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Provider  string `json:"provider"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".functionfly", "credentials.json"), nil
}

// LoadCredentials reads credentials from disk.
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in\n   → Run: fly login")
		}
		return nil, fmt.Errorf("could not read credentials: %w", err)
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("credentials file is corrupted\n   → Run: fly login")
	}
	if !creds.ExpiresAt.IsZero() && time.Now().After(creds.ExpiresAt) {
		return nil, fmt.Errorf("your session has expired\n   → Run: fly login")
	}
	return &creds, nil
}

// SaveCredentials writes credentials to disk.
func SaveCredentials(creds *Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("could not create credentials directory: %w", err)
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize credentials: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// DeleteCredentials removes the credentials file.
func DeleteCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove credentials: %w", err)
	}
	return nil
}

// IsLoggedIn returns true if valid credentials exist.
func IsLoggedIn() bool {
	_, err := LoadCredentials()
	return err == nil
}
