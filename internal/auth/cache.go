package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type fileCache struct{}

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

func getTenantPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".azsh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "tenant.json"), nil
}

func readCachedTenant() (string, error) {
	path, err := getTenantPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var payload struct {
		TenantID string `json:"tenantId"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	return payload.TenantID, nil
}

func writeCachedTenant(tenant string) error {
	if tenant == "" {
		return nil
	}

	path, err := getTenantPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(struct {
		TenantID string `json:"tenantId"`
	}{TenantID: tenant})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (fileCache) Replace(ctx context.Context, cache cache.Unmarshaler, hints cache.ReplaceHints) error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return cache.Unmarshal(data)
}

func (fileCache) Export(ctx context.Context, cache cache.Marshaler, hints cache.ExportHints) error {
	data, err := cache.Marshal()
	if err != nil {
		return err
	}

	path, err := getCachePath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
