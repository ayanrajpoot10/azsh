package auth

import (
	"context"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type tokenCache struct{}

func cachePath(filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".azsh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

func (tokenCache) Replace(ctx context.Context, c cache.Unmarshaler, hints cache.ReplaceHints) error {
	path, err := cachePath("token.json")
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

	path, err := cachePath("token.json")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
