package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ayanrajpoot10/azsh/internal/arm"
	"github.com/ayanrajpoot10/azsh/internal/utils"
)

type Settings struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	PreferredOsType    string          `json:"preferredOsType"`
	PreferredLocation  string          `json:"preferredLocation"`
	PreferredShellType string          `json:"preferredShellType"`
	NetworkType        string          `json:"networkType"`
	SessionType        string          `json:"sessionType"`
	StorageProfile     *StorageProfile `json:"storageProfile,omitempty"`
}

type StorageProfile struct {
	StorageAccountResourceID string `json:"storageAccountResourceId"`
	FileShareName            string `json:"fileShareName"`
	DiskSizeInGB             int    `json:"diskSizeInGB"`
}

func readCachedSettings() (*Properties, error) {
	path, err := utils.CachePath("settings.json")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	settings := &Settings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	return &settings.Properties, nil
}

func writeCachedSettings(props *Properties) error {
	path, err := utils.CachePath("settings.json")
	if err != nil {
		return err
	}
	data, err := json.Marshal(Settings{Properties: *props})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func GetUserSettings(token string) (*Properties, error) {
	if props, err := readCachedSettings(); err == nil {
		return props, nil
	}

	req, err := http.NewRequest(http.MethodGet, userSettingsURL, nil)
	if err != nil {
		return nil, err
	}

	arm.SetCommonHeaders(req, token)

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if err := arm.CheckStatus(resp.StatusCode); err != nil {
		return nil, fmt.Errorf("user settings: %s, response: %s", resp.Status, string(data))
	}

	settings := &Settings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	writeCachedSettings(&settings.Properties)

	return &settings.Properties, nil
}

func DeleteUserSettings(token string) error {
	req, err := http.NewRequest(http.MethodDelete, userSettingsURL, nil)
	if err != nil {
		return err
	}

	arm.SetCommonHeaders(req, token)

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := arm.CheckStatus(resp.StatusCode, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete user settings: %s, response: %s", resp.Status, string(data))
	}

	return nil
}

func IsUserSettingsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UserSettingsNotFound")
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

	arm.SetCommonHeaders(req, token)
	arm.SetContentTypeJSON(req)

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := arm.CheckStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
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
