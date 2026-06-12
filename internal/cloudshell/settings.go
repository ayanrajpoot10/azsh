package cloudshell

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

func settingsCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".azsh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

func readCachedSettings() (*Properties, error) {
	path, err := settingsCachePath()
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
	path, err := settingsCachePath()
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

	writeCachedSettings(&settings.Properties)

	return &settings.Properties, nil
}

func IsUserSettingsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UserSettingsNotFound")
}
