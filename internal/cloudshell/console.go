package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type ConsoleResponse struct {
	Properties ConsoleProperties `json:"properties"`
}

type ConsoleProperties struct {
	OsType            string `json:"osType"`
	ProvisioningState string `json:"provisioningState"`
	URI               string `json:"uri"`
}

type TerminalResponse struct {
	ID           string `json:"id"`
	SocketURI    string `json:"socketUri"`
	IdleTimeout  string `json:"idleTimeout"`
	TokenUpdated bool   `json:"tokenUpdated"`
}

func consoleCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".azsh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "console.json"), nil
}

func readCachedConsole() (*ConsoleResponse, error) {
	path, err := consoleCachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cr ConsoleResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

func writeCachedConsole(cr *ConsoleResponse) error {
	path, err := consoleCachePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func ProvisionConsole(token, osType, preferredLocation string) (*ConsoleResponse, error) {
	if cr, err := readCachedConsole(); err == nil && cr.Properties.OsType == osType {
		authURI := cr.Properties.URI + "/authorize"
		req, _ := http.NewRequest(http.MethodPost, authURI, bytes.NewBufferString("{}"))
		setCommonHeaders(req, token)
		setContentTypeJSON(req)
		resp, _, authErr := executeRequest(req)
		if authErr == nil && checkStatus(resp.StatusCode) == nil {
			return cr, nil
		}
		if path, err := consoleCachePath(); err == nil {
			os.Remove(path)
		}
	}

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
		return nil, fmt.Errorf("console provisioning: %s, response: %s", resp.Status, string(data))
	}

	var consoleResp ConsoleResponse
	if err := json.Unmarshal(data, &consoleResp); err != nil {
		return nil, err
	}

	authURI := consoleResp.Properties.URI + "/authorize"
	authReq, _ := http.NewRequest(http.MethodPost, authURI, bytes.NewBufferString("{}"))
	setCommonHeaders(authReq, token)
	setContentTypeJSON(authReq)
	authResp, _, authErr := executeRequest(authReq)
	if authErr != nil {
		return nil, fmt.Errorf("authorize console: %w", authErr)
	}
	if err := checkStatus(authResp.StatusCode); err != nil {
		return nil, fmt.Errorf("authorize console: %s", authResp.Status)
	}

	writeCachedConsole(&consoleResp)

	return &consoleResp, nil
}

func NegotiateTerminal(token, consoleURI, shell string, cols, rows int) (*TerminalResponse, error) {
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
		return nil, fmt.Errorf("negotiate terminal: %s, response: %s", resp.Status, string(data))
	}

	var terminal TerminalResponse
	if err := json.Unmarshal(data, &terminal); err != nil {
		return nil, err
	}

	return &terminal, nil
}

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
		return fmt.Errorf("resize terminal: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
