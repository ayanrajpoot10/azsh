package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

type Tenant struct {
	ID string `json:"tenantId"`
}

type TenantsResponse struct {
	Value []Tenant `json:"value"`
}

var httpClient = &http.Client{}

func getTenant(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://management.azure.com/tenants?api-version=2020-01-01", nil)
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

func refreshToken(refreshTokenStr, targetTenant string) (*TokenResponse, error) {
	uri := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", targetTenant)
	data := url.Values{}
	data.Set("client_id", defaultClientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshTokenStr)
	data.Set("scope", defaultScope)

	req, _ := http.NewRequest("POST", uri, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh token, status: %s", resp.Status)
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &tr, nil
}

func deviceLogin() (*TokenResponse, error) {
	uri := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/devicecode", defaultTenantID)
	data := url.Values{}
	data.Set("client_id", defaultClientID)
	data.Set("scope", defaultScope)

	req, _ := http.NewRequest("POST", uri, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var dc DeviceCodeResponse
	if err := json.Unmarshal(body, &dc); err != nil {
		return nil, err
	}

	fmt.Println(dc.Message)

	tokenURI := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", defaultTenantID)
	tokenData := url.Values{}
	tokenData.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	tokenData.Set("client_id", defaultClientID)
	tokenData.Set("device_code", dc.DeviceCode)

	interval := time.Duration(dc.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	return pollDeviceToken(tokenURI, tokenData, interval)
}

func pollDeviceToken(tokenURI string, tokenData url.Values, interval time.Duration) (*TokenResponse, error) {
	for {
		time.Sleep(interval)
		treq, _ := http.NewRequest("POST", tokenURI, strings.NewReader(tokenData.Encode()))
		treq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		tresp, err := httpClient.Do(treq)
		if err != nil {
			continue
		}

		tbody, _ := io.ReadAll(tresp.Body)
		tresp.Body.Close()

		var tr TokenResponse
		if err := json.Unmarshal(tbody, &tr); err != nil {
			return nil, err
		}

		if tr.Error == "authorization_pending" {
			continue
		} else if tr.Error != "" {
			return nil, fmt.Errorf("auth error: %s", tr.Error)
		}

		if tr.AccessToken != "" {
			fmt.Println("Successfully authenticated!")
			return &tr, nil
		}
	}
}
