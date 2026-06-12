package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
	DisplayName    string `json:"displayName"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type registrationPayload struct {
	Properties registrationProperties `json:"properties"`
}

type registrationProperties struct {
	PreferredOsType    string       `json:"preferredOsType"`
	PreferredLocation  string       `json:"preferredLocation"`
	StorageProfile     *string      `json:"storageProfile"`
	TerminalSettings   termSettings `json:"terminalSettings"`
	VnetSettings       *string      `json:"vnetSettings"`
	UserSubscription   string       `json:"userSubscription"`
	SessionType        string       `json:"sessionType"`
	NetworkType        string       `json:"networkType"`
	PreferredShellType string       `json:"preferredShellType"`
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

func RegisterUserSettings(token, subscriptionID, location string) error {
	payload := registrationPayload{
		Properties: registrationProperties{
			PreferredOsType:    "linux",
			PreferredLocation:  location,
			StorageProfile:     nil,
			TerminalSettings: termSettings{
				FontSize:  "medium",
				FontStyle: "monospace",
			},
			VnetSettings:       nil,
			UserSubscription:   subscriptionID,
			SessionType:        "Ephemeral",
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
	})

	return nil
}
