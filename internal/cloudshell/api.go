package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	userSettingsURL = "https://management.azure.com/providers/Microsoft.Portal/userSettings/cloudconsole?api-version=2023-02-01-preview"
	userAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
)

var client = &http.Client{}

type CloudShellSettings struct {
	Properties CloudShellProperties `json:"properties"`
}

type CloudShellProperties struct {
	PreferredOsType    string `json:"preferredOsType"`
	PreferredLocation  string `json:"preferredLocation"`
	PreferredShellType string `json:"preferredShellType"`
	NetworkType        string `json:"networkType"`
	SessionType        string `json:"sessionType"`
}

func UserSettings(token string) (*CloudShellProperties, error) {
	req, err := http.NewRequest("GET", userSettingsURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("origin", "https://ux.console.azure.com")
	req.Header.Set("referer", "https://ux.console.azure.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s, response: %s", resp.Status, string(data))
	}

	settings := &CloudShellSettings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	return &settings.Properties, nil
}

type ConsoleResponse struct {
	Properties ConsoleProperties `json:"properties"`
}

type ConsoleProperties struct {
	OsType            string `json:"osType"`
	ProvisioningState string `json:"provisioningState"`
	URI               string `json:"uri"`
}

func ProvisionConsole(token, osType, preferredLocation string) (*ConsoleResponse, error) {
	uri := "https://management.azure.com/providers/Microsoft.Portal/consoles/default?api-version=2023-02-01-preview"
	payload := fmt.Sprintf(`{"properties":{"osType":"%s"}}`, osType)
	req, _ := http.NewRequest("PUT", uri, bytes.NewBufferString(payload))

	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("origin", "https://ux.console.azure.com")
	req.Header.Set("referer", "https://ux.console.azure.com/")
	if preferredLocation != "" {
		req.Header.Set("x-ms-console-preferred-location", preferredLocation)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("console provisioning failed with status: %s, response: %s", resp.Status, string(data))
	}

	var consoleResp ConsoleResponse
	if err := json.Unmarshal(data, &consoleResp); err != nil {
		return nil, err
	}

	return &consoleResp, nil
}

type TerminalResponse struct {
	ID           string `json:"id"`
	SocketURI    string `json:"socketUri"`
	IdleTimeout  string `json:"idleTimeout"`
	TokenUpdated bool   `json:"tokenUpdated"`
}

func NegotiateTerminal(token, consoleURI, shell string, cols, rows int) (*TerminalResponse, error) {
	authURI := fmt.Sprintf("%s/authorize", consoleURI)
	authReq, _ := http.NewRequest("POST", authURI, bytes.NewBufferString("{}"))

	authReq.Header.Set("authorization", "Bearer "+token)
	authReq.Header.Set("accept", "application/json")
	authReq.Header.Set("content-type", "application/json")
	authReq.Header.Set("user-agent", userAgent)
	authReq.Header.Set("origin", "https://ux.console.azure.com")
	authReq.Header.Set("referer", "https://ux.console.azure.com/")

	authResp, err := client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer authResp.Body.Close()

	if authResp.StatusCode < 200 || authResp.StatusCode >= 300 {
		b, _ := io.ReadAll(authResp.Body)
		return nil, fmt.Errorf("authorization to container failed: %s - %s", authResp.Status, string(b))
	}

	termURI := fmt.Sprintf("%s/terminals?cols=%d&rows=%d&version=2019-01-01&shell=%s", consoleURI, cols, rows, shell)
	termReq, _ := http.NewRequest("POST", termURI, bytes.NewBufferString("{}"))

	termReq.Header.Set("authorization", "Bearer "+token)
	termReq.Header.Set("accept", "application/json")
	termReq.Header.Set("content-type", "application/json")
	termReq.Header.Set("user-agent", userAgent)
	termReq.Header.Set("origin", "https://ux.console.azure.com")
	termReq.Header.Set("referer", "https://ux.console.azure.com/")

	termResp, err := client.Do(termReq)
	if err != nil {
		return nil, err
	}
	defer termResp.Body.Close()

	data, err := io.ReadAll(termResp.Body)
	if err != nil {
		return nil, err
	}

	if termResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("terminal negotiation failed with status: %s, response: %s", termResp.Status, string(data))
	}

	var terminal TerminalResponse
	if err := json.Unmarshal(data, &terminal); err != nil {
		return nil, err
	}

	return &terminal, nil
}

func ResizeTerminal(token, consoleURI, terminalID string, cols, rows int) error {
	uri := fmt.Sprintf("%s/terminals/%s/size?cols=%d&rows=%d&version=2019-01-01", consoleURI, terminalID, cols, rows)
	req, _ := http.NewRequest("POST", uri, bytes.NewBufferString("{}"))

	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("origin", "https://ux.console.azure.com")
	req.Header.Set("referer", "https://ux.console.azure.com/")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("terminal resize failed with status: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
