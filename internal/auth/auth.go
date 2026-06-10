package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	"github.com/ayanrajpoot10/azsh/internal/utils"
)

const (
	clientID        = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	defaultTenantID = "organizations"
	msLoginBase     = "https://login.microsoftonline.com"
	defaultScope    = "https://management.core.windows.net//.default"
)

var httpClient = &http.Client{}

type tenant struct {
	ID string `json:"tenantId"`
}

type tenantsResponse struct {
	Value []tenant `json:"value"`
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

	tenant, err := selectTenant(token)
	if err != nil {
		return "", fmt.Errorf("tenant selection: %w", err)
	}

	token, err = tokenForTenant(tenant)
	if err != nil {
		return "", fmt.Errorf("tenant token: %w", err)
	}

	return token, nil
}

func silentAuth() (string, error) {
	ctx := context.Background()
	client, err := public.New(
		clientID,
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
		public.WithInstanceDiscovery(false),
	)
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

	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, defaultTenantID)),
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
	)
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

func tokenForTenant(tenant string) (string, error) {
	ctx := context.Background()

	client, err := public.New(
		clientID,
		public.WithAuthority(fmt.Sprintf("%s/%s", msLoginBase, tenant)),
		public.WithCache(tokenCache{}),
		public.WithHTTPClient(httpClient),
		public.WithInstanceDiscovery(false),
	)
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
			public.WithTenantID(tenant),
		)
		if err == nil {
			return result.AccessToken, nil
		}
	}

	return "", fmt.Errorf("token for tenant %s", tenant)
}

func selectTenant(token string) (string, error) {
	tenants, err := getTenants(token)
	if err != nil {
		return "", err
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

func getTenants(token string) ([]tenant, error) {
	req, err := http.NewRequest(http.MethodGet, "https://management.azure.com/tenants?api-version=2020-01-01", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tenants: %s", resp.Status)
	}

	var tr tenantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	if len(tr.Value) == 0 {
		return nil, fmt.Errorf("no tenants found")
	}

	return tr.Value, nil
}
