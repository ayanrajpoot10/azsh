package cloudshell

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Settings struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	PreferredOsType    string `json:"preferredOsType"`
	PreferredLocation  string `json:"preferredLocation"`
	PreferredShellType string `json:"preferredShellType"`
	NetworkType        string `json:"networkType"`
	SessionType        string `json:"sessionType"`
}

func GetUserSettings(token string) (*Properties, error) {
	req, err := http.NewRequest(http.MethodGet, userSettingsURL, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, token)

	resp, data, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
		return nil, fmt.Errorf("user settings: %s, response: %s", resp.Status, string(data))
	}

	settings := &Settings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	return &settings.Properties, nil
}

func IsUserSettingsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UserSettingsNotFound")
}
