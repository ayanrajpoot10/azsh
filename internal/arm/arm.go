package arm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ListTenants(token string) ([]Tenant, error) {
	req, err := http.NewRequest(http.MethodGet, tenantsURL, nil)
	if err != nil {
		return nil, err
	}

	SetCommonHeaders(req, token)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list tenants: %s, response: %s", resp.Status, string(data))
	}

	var result struct {
		Value []Tenant `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

func ListSubscriptions(token string) ([]Subscription, error) {
	req, err := http.NewRequest(http.MethodGet, subscriptionsURL, nil)
	if err != nil {
		return nil, err
	}

	SetCommonHeaders(req, token)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list subscriptions: %s, response: %s", resp.Status, string(data))
	}

	var result struct {
		Value []Subscription `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

func ListResourceGroups(token, subscriptionID string) ([]ResourceGroup, error) {
	url := fmt.Sprintf(resourceGroupsURL, subscriptionID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	SetCommonHeaders(req, token)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list resource groups: %s, response: %s", resp.Status, string(data))
	}

	var result struct {
		Value []ResourceGroup `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

func ListStorageAccounts(token, subscriptionID, resourceGroup string) ([]StorageAccountInfo, error) {
	url := fmt.Sprintf(listStorageAccountsURL, subscriptionID, resourceGroup)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	SetCommonHeaders(req, token)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list storage accounts: %s, response: %s", resp.Status, string(data))
	}

	var result struct {
		Value []StorageAccountInfo `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}
