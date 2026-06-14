package arm

import (
	"fmt"
	"io"
	"net/http"
)

const (
	subscriptionsURL       = "https://management.azure.com/subscriptions?api-version=2018-07-01"
	resourceGroupsURL      = "https://management.azure.com/subscriptions/%s/resourceGroups?api-version=2017-05-10"
	resourceGroupURL       = "https://management.azure.com/subscriptions/%s/resourceGroups/%s?api-version=2018-07-01"
	listStorageAccountsURL = "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts?api-version=2022-09-01"
	storageAccountURL      = "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s?api-version=2022-09-01"
	registerRPURL          = "https://management.azure.com/subscriptions/%s/providers/Microsoft.CloudShell?api-version=2022-12-01"
	consoleOrigin          = "https://ux.console.azure.com"
	userAgent              = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
)

var client = &http.Client{}

func SetCommonHeaders(req *http.Request, token string) {
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("origin", consoleOrigin)
	req.Header.Set("referer", consoleOrigin+"/")
}

func SetContentTypeJSON(req *http.Request) {
	req.Header.Set("content-type", "application/json")
}

func ExecuteRequest(req *http.Request) (*http.Response, []byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	return resp, data, nil
}

