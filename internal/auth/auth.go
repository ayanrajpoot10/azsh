package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	clientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	tenantID = "organizations"
)

var defaultScopes = []string{"https://management.core.windows.net//.default"}

func Auth() (string, error) {
	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID)),
		public.WithCache(fileCache{}),
		public.WithHTTPClient(httpClient),
	)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		opts := []public.AcquireSilentOption{public.WithSilentAccount(account)}
		opts = append(opts, public.WithTenantID(tenantID))

		result, err := client.AcquireTokenSilent(ctx, defaultScopes, opts...)
		if err != nil {
			continue
		}

		return result.AccessToken, nil
	}

	dc, err := client.AcquireTokenByDeviceCode(ctx, defaultScopes, public.WithTenantID(tenantID))
	if err != nil {
		return "", err
	}

	log.Println(dc.Result.Message)

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

func RefreshTokenWithTenant(tenant string) (string, error) {
	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("https://login.microsoftonline.com/%s", tenant)),
		public.WithCache(fileCache{}),
		public.WithHTTPClient(httpClient),
	)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		opts := []public.AcquireSilentOption{public.WithSilentAccount(account)}
		opts = append(opts, public.WithTenantID(tenant))

		result, err := client.AcquireTokenSilent(ctx, defaultScopes, opts...)
		if err != nil {
			continue
		}

		return result.AccessToken, nil
	}

	return "", fmt.Errorf("failed to acquire tenant token for tenant %s", tenant)
}

func GetToken() (string, error) {
	if tenant, err := readCachedTenant(); err == nil && tenant != "" {
		if token, err := RefreshTokenWithTenant(tenant); err == nil {
			return token, nil
		}
	}

	token, err := Auth()
	if err != nil {
		return "", err
	}

	tenant, err := GetTenant(token)
	if err != nil {
		return "", err
	}

	if tenant == "" {
		return "", fmt.Errorf("no tenant found")
	}

	if err := writeCachedTenant(tenant); err != nil {
		return "", err
	}

	token, err = RefreshTokenWithTenant(tenant)
	if err != nil {
		return "", err
	}

	return token, nil
}
