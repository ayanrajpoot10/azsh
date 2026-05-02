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

// GetTenant fetches the tenant ID, with interactive selection if multiple tenants are available
func GetTenant(token string) (string, error) {
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

	return selectTenant(tenants)
}

func selectTenant(tenants []Tenant) (string, error) {
	fmt.Println("\nMultiple tenants found. Please select one:")
	fmt.Println()

	for i, tenant := range tenants {
		fmt.Printf("  [%d] %s\n", i+1, tenant.ID)
	}

	fmt.Println()

	var choice int
	for {
		fmt.Printf("Enter your choice (1-%d): ", len(tenants))
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(tenants) {
			fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(tenants))
			continue
		}
		break
	}

	return tenants[choice-1].ID, nil
}
