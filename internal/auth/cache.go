package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type TokenCache struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TenantID     string `json:"tenant_id"`
}

func (t *TokenCache) IsExpired() bool {
	return time.Now().Unix() >= t.ExpiresAt
}

func getCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".azsh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

func loadCache() (*TokenCache, error) {
	path, err := getCachePath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache TokenCache
	if err := json.Unmarshal(b, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func saveCache(tr *TokenResponse, tenantID string) error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	cache := TokenCache{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Now().Unix() + int64(tr.ExpiresIn) - 300,
		TenantID:     tenantID,
	}

	b, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0600)
}
