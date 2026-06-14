package auth

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

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

	dc, err := client.AcquireTokenByDeviceCode(ctx, []string{defaultScope}, public.WithTenantID(defaultTenantID))
	if err != nil {
		return "", err
	}

	fmt.Println()
	fmt.Println("To sign in, open:")
	fmt.Printf("  %s\n", dc.Result.VerificationURL)
	fmt.Println()
	fmt.Println("And enter the code:")
	fmt.Printf("  %s\n", dc.Result.UserCode)
	fmt.Println()

	fmt.Print("Press Enter to copy code and open browser...")
	bufio.NewScanner(os.Stdin).Scan()

	if err := clipboard.WriteAll(dc.Result.UserCode); err != nil {
		fmt.Println()
		fmt.Printf("Warning: failed to copy code to clipboard: %v\n", err)
	}
	if err := browser.OpenURL(dc.Result.VerificationURL); err != nil {
		fmt.Printf("Warning: failed to open browser: %v\n", err)
	}
	fmt.Println()
	fmt.Println("✓ Device code copied! Opening browser...")
	fmt.Println()

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
