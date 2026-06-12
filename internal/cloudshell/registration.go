package cloudshell

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
	DisplayName    string `json:"displayName"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type StorageAccountInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Sku      struct {
		Name string `json:"name"`
	} `json:"sku"`
	Properties struct {
		ProvisioningState string `json:"provisioningState"`
	} `json:"properties"`
}

type StorageProfile struct {
	StorageAccountResourceID string `json:"storageAccountResourceId"`
	FileShareName            string `json:"fileShareName"`
	DiskSizeInGB             int    `json:"diskSizeInGB"`
}

type storageAccountCreatePayload struct {
	Location   string          `json:"location"`
	Sku        map[string]string `json:"sku"`
	Kind       string          `json:"kind"`
	Tags       map[string]string `json:"tags"`
	Properties accountProps     `json:"properties"`
}

type accountProps struct {
	Encryption                encryptionConfig `json:"encryption"`
	SupportsHTTPSOnly         bool             `json:"supportsHttpsTrafficOnly"`
	AllowBlobPublicAccess     bool             `json:"allowBlobPublicAccess"`
	MinimumTLSVersion         string           `json:"minimumTlsVersion"`
}

type encryptionConfig struct {
	Services   encryptionServices `json:"services"`
	KeySource  string             `json:"keySource"`
}

type encryptionServices struct {
	Blob serviceConfig `json:"blob"`
	File serviceConfig `json:"file"`
}

type serviceConfig struct {
	Enabled bool `json:"enabled"`
}

type registrationPayload struct {
	Properties registrationProperties `json:"properties"`
}

type registrationProperties struct {
	PreferredOsType    string          `json:"preferredOsType"`
	PreferredLocation  string          `json:"preferredLocation"`
	StorageProfile     *StorageProfile `json:"storageProfile"`
	TerminalSettings   termSettings    `json:"terminalSettings"`
	VnetSettings       *string         `json:"vnetSettings"`
	UserSubscription   string          `json:"userSubscription"`
	SessionType        string          `json:"sessionType"`
	NetworkType        string          `json:"networkType"`
	PreferredShellType string          `json:"preferredShellType"`
}

type termSettings struct {
	FontSize  string `json:"fontSize"`
	FontStyle string `json:"fontStyle"`
}

func ListSubscriptions(token string) ([]Subscription, error) {
	req, err := http.NewRequest(http.MethodGet, subscriptionsURL, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, token)

	resp, data, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
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

	setCommonHeaders(req, token)

	resp, data, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
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

	setCommonHeaders(req, token)

	resp, data, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
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

func CreateResourceGroup(token, subscriptionID, name, location string) error {
	url := fmt.Sprintf(resourceGroupURL, subscriptionID, name)
	body := fmt.Sprintf(`{"location":"%s"}`, location)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	setCommonHeaders(req, token)
	setContentTypeJSON(req)

	resp, data, err := executeRequest(req)
	if err != nil {
		return err
	}

	if err := checkStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
		return fmt.Errorf("create resource group: %s, response: %s", resp.Status, string(data))
	}

	return nil
}

func CreateStorageAccount(token, subscriptionID, resourceGroup, accountName, location string) error {
	url := fmt.Sprintf(storageAccountURL, subscriptionID, resourceGroup, accountName)

	payload := storageAccountCreatePayload{
		Location: location,
		Sku:      map[string]string{"name": "Standard_LRS"},
		Kind:     "StorageV2",
		Tags:     map[string]string{"ms-resource-usage": "azure-cloud-shell"},
		Properties: accountProps{
			Encryption: encryptionConfig{
				Services: encryptionServices{
					Blob: serviceConfig{Enabled: true},
					File: serviceConfig{Enabled: true},
				},
				KeySource: "Microsoft.Storage",
			},
			SupportsHTTPSOnly:     true,
			AllowBlobPublicAccess: false,
			MinimumTLSVersion:     "TLS1_2",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	setCommonHeaders(req, token)
	setContentTypeJSON(req)

	resp, data, err := executeRequest(req)
	if err != nil {
		return err
	}

	// 202 = accepted (async), 200 = already exists/completed
	if err := checkStatus(resp.StatusCode, http.StatusOK, http.StatusAccepted); err != nil {
		return fmt.Errorf("create storage account: %s, response: %s", resp.Status, string(data))
	}

	if resp.StatusCode == http.StatusAccepted {
		return pollStorageAccount(token, subscriptionID, resourceGroup, accountName)
	}

	return nil
}

func pollStorageAccount(token, subscriptionID, resourceGroup, accountName string) error {
	url := fmt.Sprintf(storageAccountURL, subscriptionID, resourceGroup, accountName)

	for i := 0; i < 60; i++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		setCommonHeaders(req, token)

		resp, data, err := executeRequest(req)
		if err != nil {
			return err
		}

		if err := checkStatus(resp.StatusCode); err != nil {
			return fmt.Errorf("poll storage account: %s, response: %s", resp.Status, string(data))
		}

		var account StorageAccountInfo
		if err := json.Unmarshal(data, &account); err != nil {
			return err
		}

		if account.Properties.ProvisioningState == "Succeeded" {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("storage account provisioning timed out")
}

func RegisterCloudShellRP(token, subscriptionID string) error {
	url := fmt.Sprintf(registerRPURL, subscriptionID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	setCommonHeaders(req, token)

	resp, data, err := executeRequest(req)
	if err != nil {
		return err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
		return fmt.Errorf("register RP: %s, response: %s", resp.Status, string(data))
	}

	return nil
}

func RegisterUserSettings(token, subscriptionID, location string, storageProfile *StorageProfile) error {
	sessionType := "Ephemeral"
	if storageProfile != nil {
		sessionType = "Mounted"
	}

	payload := registrationPayload{
		Properties: registrationProperties{
			PreferredOsType:    "linux",
			PreferredLocation:  location,
			StorageProfile:     storageProfile,
			TerminalSettings: termSettings{
				FontSize:  "medium",
				FontStyle: "monospace",
			},
			VnetSettings:       nil,
			UserSubscription:   subscriptionID,
			SessionType:        sessionType,
			NetworkType:        "Default",
			PreferredShellType: "bash",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, userSettingsURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	setCommonHeaders(req, token)
	setContentTypeJSON(req)

	resp, data, err := executeRequest(req)
	if err != nil {
		return err
	}

	if err := checkStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
		return fmt.Errorf("user settings registration failed: %s, response: %s", resp.Status, string(data))
	}

	writeCachedSettings(&Properties{
		PreferredOsType:    payload.Properties.PreferredOsType,
		PreferredLocation:  payload.Properties.PreferredLocation,
		PreferredShellType: payload.Properties.PreferredShellType,
		NetworkType:        payload.Properties.NetworkType,
		SessionType:        payload.Properties.SessionType,
		StorageProfile:     storageProfile,
	})

	return nil
}

func GenerateStorageNames(token string) (accountName, fileShareName string) {
	puid, email := extractClaimsFromToken(token)

	if puid == "" {
		return "", ""
	}

	normalizedEmail := strings.NewReplacer(
		"@", "-",
		".", "-",
		"_", "-",
	).Replace(email)

	accountName = "csg" + puid
	if len(accountName) > 24 {
		accountName = accountName[:24]
	}

	fileShareName = "cs-" + normalizedEmail + "-" + puid

	return accountName, fileShareName
}

func extractClaimsFromToken(token string) (puid, email string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", ""
	}

	if p, ok := claims["puid"].(string); ok {
		puid = strings.ToLower(p)
	} else if oid, ok := claims["oid"].(string); ok {
		puid = strings.ToLower(strings.ReplaceAll(oid, "-", ""))
	}

	if e, ok := claims["email"].(string); ok {
		email = e
	} else if un, ok := claims["unique_name"].(string); ok {
		email = un
	}

	return puid, email
}
