package auth

import (
	"context"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type tokenCache struct{}

const (
	configDir = ".azsh"
	tokenFile = "token.json"
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
