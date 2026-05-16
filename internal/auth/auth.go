package auth

import (
	"context"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	clientID        = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	defaultTenantID = "organizations"
	msLoginBase     = "https://login.microsoftonline.com"
	defaultScope    = "https://management.core.windows.net//.default"
)

var defaultScopes = []string{defaultScope}

func newClient(tenant string) (public.Client, error) {
	return public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, tenant)),
		public.WithCache(fileCache{}),
		public.WithHTTPClient(httpClient),
	)
}

func acquireToken(ctx context.Context, tenant string) (string, error) {
	client, err := newClient(tenant)
	if err != nil {
		return "", err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		opts := []public.AcquireSilentOption{
			public.WithSilentAccount(account),
			public.WithTenantID(tenant),
		}

		result, err := client.AcquireTokenSilent(ctx, defaultScopes, opts...)
		if err != nil {
			continue
		}

		return result.AccessToken, nil
	}

	return "", fmt.Errorf("no cached token found")
}

func deviceCodeLogin() (string, error) {
	ctx := context.Background()

	token, err := acquireToken(ctx, defaultTenantID)
	if err == nil {
		return token, nil
	}

	client, err := newClient(defaultTenantID)
	if err != nil {
		return "", err
	}

	dc, err := client.AcquireTokenByDeviceCode(ctx, defaultScopes, public.WithTenantID(defaultTenantID))
	if err != nil {
		return "", err
	}

	fmt.Println(dc.Result.Message)

	deviceCtx := ctx
	if !dc.Result.ExpiresOn.IsZero() {
		var cancel context.CancelFunc
		deviceCtx, cancel = context.WithDeadline(ctx, dc.Result.ExpiresOn)
		defer cancel()
	}

	result, err := dc.AuthenticationResult(deviceCtx)
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func refreshTokenWithTenant(tenant string) (string, error) {
	ctx := context.Background()
	token, err := acquireToken(ctx, tenant)
	if err != nil {
		return "", fmt.Errorf("failed to acquire token for tenant %s: %w", tenant, err)
	}
	return token, nil
}

func Token() (string, error) {
	tenant, _ := readCachedTenant()

	if tenant != "" {
		if token, err := refreshTokenWithTenant(tenant); err == nil {
			return token, nil
		}
	}

	token, err := deviceCodeLogin()
	if err != nil {
		return "", err
	}

	newTenant, err := SelectTenant(token)
	if err != nil {
		return "", err
	}

	if err := writeCachedTenant(newTenant); err != nil {
		return "", err
	}

	token, err = refreshTokenWithTenant(newTenant)
	if err != nil {
		return "", err
	}

	return token, nil
}
