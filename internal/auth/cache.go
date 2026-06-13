package auth

import (
	"context"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"

	"github.com/ayanrajpoot10/azsh/internal/utils"
)

type tokenCache struct{}

func (tokenCache) Replace(ctx context.Context, c cache.Unmarshaler, hints cache.ReplaceHints) error {
	path, err := utils.CachePath("token.json")
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

	path, err := utils.CachePath("token.json")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
