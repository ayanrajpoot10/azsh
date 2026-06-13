package arm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func CreateResourceGroup(token, subscriptionID, name, location string) error {
	url := fmt.Sprintf(resourceGroupURL, subscriptionID, name)
	body := fmt.Sprintf(`{"location":"%s"}`, location)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	SetCommonHeaders(req, token)
	SetContentTypeJSON(req)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := CheckStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
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

	SetCommonHeaders(req, token)
	SetContentTypeJSON(req)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := CheckStatus(resp.StatusCode, http.StatusOK, http.StatusAccepted); err != nil {
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

		SetCommonHeaders(req, token)

		resp, data, err := ExecuteRequest(req)
		if err != nil {
			return err
		}

		if err := CheckStatus(resp.StatusCode); err != nil {
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

	SetCommonHeaders(req, token)

	resp, data, err := ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := CheckStatus(resp.StatusCode); err != nil {
		return fmt.Errorf("register RP: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
