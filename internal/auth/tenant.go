package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ayanrajpoot10/azsh/internal/utils"
)

type Tenant struct {
	ID string `json:"tenantId"`
}

type TenantsResponse struct {
	Value []Tenant `json:"value"`
}

var httpClient = &http.Client{}

func getTenants(token string) ([]Tenant, error) {
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
		return nil, fmt.Errorf("failed to get tenants: %s", resp.Status)
	}

	var tr TenantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	if len(tr.Value) == 0 {
		return nil, fmt.Errorf("no tenants found")
	}

	return tr.Value, nil
}

func SelectTenant(token string) (string, error) {
	tenants, err := getTenants(token)
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
