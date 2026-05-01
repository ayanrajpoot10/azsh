package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Tenant struct {
	ID string `json:"tenantId"`
}

type TenantsResponse struct {
	Value []Tenant `json:"value"`
}

var httpClient = &http.Client{}

func GetTenant(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://management.azure.com/tenants?api-version=2020-01-01", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get tenants: %s", resp.Status)
	}

	var tr TenantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}

	if len(tr.Value) == 0 {
		return "", fmt.Errorf("no tenants found")
	}

	return tr.Value[0].ID, nil
}
