package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/ayanrajpoot10/azsh/internal/arm"
	"github.com/ayanrajpoot10/azsh/internal/utils"
)

const (
	clientID        = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	defaultTenantID = "organizations"
	msLoginBase     = "https://login.microsoftonline.com"
	defaultScope    = "https://management.core.windows.net//.default"
)

var httpClient = &http.Client{}

func newClient(tenantID string) (public.Client, error) {
	opts := []public.Option{
		public.WithCache(fileCache{}),
		public.WithHTTPClient(httpClient),
		public.WithInstanceDiscovery(false),
	}
	if tenantID != "" {
		opts = append(opts, public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, tenantID)))
	}
	return public.New(clientID, opts...)
}

func Authenticate() (string, error) {
	token, err := silentAuth()
	if err == nil {
		return token, nil
	}

	token, err = interactiveLogin()
	if err != nil {
		return "", fmt.Errorf("interactive login: %w", err)
	}

	tenantID, err := selectTenant(token)
	if err != nil {
		return "", fmt.Errorf("tenant selection: %w", err)
	}

	token, err = tokenForTenant(tenantID)
	if err != nil {
		return "", fmt.Errorf("tenant token: %w", err)
	}

	return token, nil
}

func silentAuth() (string, error) {
	ctx := context.Background()
	client, err := newClient("")
	if err != nil {
		return "", err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		if account.Realm == defaultTenantID {
			continue
		}
		result, err := client.AcquireTokenSilent(ctx, []string{defaultScope},
			public.WithSilentAccount(account),
			public.WithTenantID(account.Realm),
		)
		if err == nil {
			return result.AccessToken, nil
		}
	}

	return "", fmt.Errorf("no cached token")
}

func interactiveLogin() (string, error) {
	ctx := context.Background()

	client, err := newClient(defaultTenantID)
	if err != nil {
		return "", err
	}

	result, err := client.AcquireTokenInteractive(ctx, []string{defaultScope},
		public.WithTenantID(defaultTenantID),
	)
	if err != nil {
		return "", err
	}

	fmt.Println("✓ Login successful!")
	fmt.Println()

	return result.AccessToken, nil
}

func tokenForTenant(tenantID string) (string, error) {
	ctx := context.Background()

	client, err := newClient(tenantID)
	if err != nil {
		return "", err
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		result, err := client.AcquireTokenSilent(ctx, []string{defaultScope},
			public.WithSilentAccount(account),
			public.WithTenantID(tenantID),
		)
		if err == nil {
			return result.AccessToken, nil
		}
	}

	return "", fmt.Errorf("token for tenant %s", tenantID)
}

func selectTenant(token string) (string, error) {
	tenants, err := arm.ListTenants(token)
	if err != nil {
		return "", err
	}

	if len(tenants) == 0 {
		return "", fmt.Errorf("no tenants found")
	}

	if len(tenants) == 1 {
		return tenants[0].ID, nil
	}

	options := make([]string, len(tenants))
	for i, t := range tenants {
		options[i] = t.ID
	}
	idx, err := utils.PromptSelect("\nMultiple tenants found. Please select one:", options)
	if err != nil {
		return "", err
	}
	return tenants[idx].ID, nil
}
