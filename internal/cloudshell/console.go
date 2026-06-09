package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
		return nil, fmt.Errorf("console provisioning: %s, response: %s", resp.Status, string(data))
	}

	var consoleResp ConsoleResponse
	if err := json.Unmarshal(data, &consoleResp); err != nil {
		return nil, err
	}

	return &consoleResp, nil
}

func NegotiateTerminal(token, consoleURI, shell string, cols, rows int) (*TerminalResponse, error) {
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
		return nil, fmt.Errorf("authorize container: %s - %s", authResp.Status, string(authData))
	}

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
