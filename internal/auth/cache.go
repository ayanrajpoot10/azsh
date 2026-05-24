package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type tokenCache struct{}

type tenantPayload struct {
	TenantID string `json:"tenantId"`
}

const (
	configDir  = ".azsh"
	tenantFile = "tenant.json"
	tokenFile  = "token.json"
)

func getFilePath(filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

func readCachedTenant() (string, error) {
	path, err := getFilePath(tenantFile)
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

	var payload tenantPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	return payload.TenantID, nil
}

func writeCachedTenant(tenant string) error {
	if tenant == "" {
		return nil
	}

	path, err := getFilePath(tenantFile)
	if err != nil {
		return err
	}

	data, err := json.Marshal(tenantPayload{TenantID: tenant})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (tokenCache) Replace(ctx context.Context, c cache.Unmarshaler, hints cache.ReplaceHints) error {
	path, err := getFilePath(tokenFile)
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

	return c.Unmarshal(data)
}

func (tokenCache) Export(ctx context.Context, c cache.Marshaler, hints cache.ExportHints) error {
	data, err := c.Marshal()
	if err != nil {
		return err
	}

	path, err := getFilePath(tokenFile)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
