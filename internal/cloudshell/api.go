package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	userSettingsURL = "https://management.azure.com/providers/Microsoft.Portal/userSettings/cloudconsole?api-version=2023-02-01-preview"
	consoleURL      = "https://management.azure.com/providers/Microsoft.Portal/consoles/default?api-version=2023-02-01-preview"
	consoleOrigin   = "https://ux.console.azure.com"
	userAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

	terminalVersion = "2019-01-01"
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

type ConsoleResponse struct {
	Properties ConsoleProperties `json:"properties"`
}

type ConsoleProperties struct {
	OsType            string `json:"osType"`
	ProvisioningState string `json:"provisioningState"`
	URI               string `json:"uri"`
}

type TerminalResponse struct {
	ID          string `json:"id"`
	SocketURI   string `json:"socketUri"`
	IdleTimeout string `json:"idleTimeout"`
	TokenUpdated bool   `json:"tokenUpdated"`
}

func setCommonHeaders(req *http.Request, token string) {
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("origin", consoleOrigin)
	req.Header.Set("referer", consoleOrigin+"/")
}

func setContentTypeJSON(req *http.Request) {
	req.Header.Set("content-type", "application/json")
}

func executeRequest(req *http.Request) (*http.Response, []byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, data, nil
}

func checkStatus(statusCode int, allowedCodes ...int) error {
	if len(allowedCodes) == 0 {
		allowedCodes = []int{http.StatusOK}
	}

	for _, code := range allowedCodes {
		if statusCode == code {
			return nil
		}
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}

// UserSettings fetches the cloud shell user settings
func UserSettings(token string) (*CloudShellProperties, error) {
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
		return nil, fmt.Errorf("user settings request failed with status: %s, response: %s", resp.Status, string(data))
	}

	settings := &CloudShellSettings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	return &settings.Properties, nil
}

// ProvisionConsole creates a new cloud shell console
func ProvisionConsole(token, osType, preferredLocation string) (*ConsoleResponse, error) {
	payload := fmt.Sprintf(`{"properties":{"osType":"%s"}}`, osType)
	req, err := http.NewRequest(http.MethodPut, consoleURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, token)
	setContentTypeJSON(req)
	if preferredLocation != "" {
		req.Header.Set("x-ms-console-preferred-location", preferredLocation)
	}

	resp, data, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("console provisioning failed: %s, response: %s", resp.Status, string(data))
	}

	var consoleResp ConsoleResponse
	if err := json.Unmarshal(data, &consoleResp); err != nil {
		return nil, err
	}

	return &consoleResp, nil
}

// NegotiateTerminal authorizes and creates a new terminal
func NegotiateTerminal(token, consoleURI, shell string, cols, rows int) (*TerminalResponse, error) {
	// Authorize access to console
	authURI := fmt.Sprintf("%s/authorize", consoleURI)
	authReq, err := http.NewRequest(http.MethodPost, authURI, bytes.NewBufferString("{}"))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(authReq, token)
	setContentTypeJSON(authReq)

	authResp, authData, err := executeRequest(authReq)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(authResp.StatusCode); err != nil {
		return nil, fmt.Errorf("authorization to container failed: %s - %s", authResp.Status, string(authData))
	}

	// Create terminal
	termURI := fmt.Sprintf("%s/terminals?cols=%d&rows=%d&version=%s&shell=%s", consoleURI, cols, rows, terminalVersion, shell)
	termReq, err := http.NewRequest(http.MethodPost, termURI, bytes.NewBufferString("{}"))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(termReq, token)
	setContentTypeJSON(termReq)

	resp, data, err := executeRequest(termReq)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
		return nil, fmt.Errorf("terminal negotiation failed: %s, response: %s", resp.Status, string(data))
	}

	var terminal TerminalResponse
	if err := json.Unmarshal(data, &terminal); err != nil {
		return nil, err
	}

	return &terminal, nil
}

// ResizeTerminal resizes the terminal to the specified dimensions
func ResizeTerminal(token, consoleURI, terminalID string, cols, rows int) error {
	uri := fmt.Sprintf("%s/terminals/%s/size?cols=%d&rows=%d&version=%s", consoleURI, terminalID, cols, rows, terminalVersion)
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	setCommonHeaders(req, token)
	setContentTypeJSON(req)

	resp, data, err := executeRequest(req)
	if err != nil {
		return err
	}

	if err := checkStatus(resp.StatusCode); err != nil {
		return fmt.Errorf("terminal resize failed: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
