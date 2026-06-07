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

func trySilentAuth() (string, error) {
	ctx := context.Background()
	client, err := public.New(
		clientID,
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
	)
	if err != nil {
		return "", err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		result, err := client.AcquireTokenSilent(ctx, defaultScopes,
			public.WithSilentAccount(account),
			public.WithTenantID(account.Realm),
		)
		if err == nil {
			return result.AccessToken, nil
		}
	}

	return "", fmt.Errorf("no cached token found")
}

func deviceCodeLogin() (string, error) {
	ctx := context.Background()

	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, defaultTenantID)),
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
	)
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

func refreshForTenant(tenant string) (string, error) {
	ctx := context.Background()

	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, tenant)),
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
	)
	if err != nil {
		return "", err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		result, err := client.AcquireTokenSilent(ctx, defaultScopes,
			public.WithSilentAccount(account),
			public.WithTenantID(tenant),
		)
		if err == nil {
			return result.AccessToken, nil
		}
	}

	return "", fmt.Errorf("failed to acquire token for tenant %s", tenant)
}

func Authenticate() (string, error) {
	token, err := trySilentAuth()
	if err == nil {
		return token, nil
	}

	token, err = deviceCodeLogin()
	if err != nil {
		return "", err
	}

	newTenant, err := SelectTenant(token)
	if err != nil {
		return "", err
	}

	token, err = refreshForTenant(newTenant)
	if err != nil {
		return "", err
	}

	if err := cleanupDefaultAccount(newTenant); err != nil {
		return "", err
	}

	return token, nil
}

func cleanupDefaultAccount(selectedTenant string) error {
	ctx := context.Background()

	client, err := public.New(
		clientID,
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
	)
	if err != nil {
		return err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		if acc.Realm != selectedTenant {
			if err := client.RemoveAccount(ctx, acc); err != nil {
				return err
			}
		}
	}

	return nil
}
