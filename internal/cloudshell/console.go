package cloudshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ayanrajpoot10/azsh/internal/arm"
	"github.com/ayanrajpoot10/azsh/internal/utils"
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

func readCachedConsole() (*ConsoleResponse, error) {
	path, err := utils.CachePath("console.json")
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
	path, err := utils.CachePath("console.json")
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
		arm.SetCommonHeaders(req, token)
		arm.SetContentTypeJSON(req)
		resp, _, authErr := arm.ExecuteRequest(req)
		if authErr == nil && arm.CheckStatus(resp.StatusCode) == nil {
			return cr, nil
		}
		if path, err := utils.CachePath("console.json"); err == nil {
			os.Remove(path)
		}
	}

	payload := fmt.Sprintf(`{"properties":{"osType":"%s"}}`, osType)
	req, err := http.NewRequest(http.MethodPut, consoleURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}

	arm.SetCommonHeaders(req, token)
	arm.SetContentTypeJSON(req)
	if preferredLocation != "" {
		req.Header.Set("x-ms-console-preferred-location", preferredLocation)
	}

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	if err := arm.CheckStatus(resp.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("console provisioning: %s, response: %s", resp.Status, string(data))
	}

	var consoleResp ConsoleResponse
	if err := json.Unmarshal(data, &consoleResp); err != nil {
		return nil, err
	}

	authURI := consoleResp.Properties.URI + "/authorize"
	authReq, _ := http.NewRequest(http.MethodPost, authURI, bytes.NewBufferString("{}"))
	arm.SetCommonHeaders(authReq, token)
	arm.SetContentTypeJSON(authReq)
	authResp, _, authErr := arm.ExecuteRequest(authReq)
	if authErr != nil {
		return nil, fmt.Errorf("authorize console: %w", authErr)
	}
	if err := arm.CheckStatus(authResp.StatusCode); err != nil {
		return nil, fmt.Errorf("authorize console: %s", authResp.Status)
	}

	writeCachedConsole(&consoleResp)

	return &consoleResp, nil
}

func CreateTerminal(token, consoleURI, shell string, cols, rows int) (*TerminalResponse, error) {
	termURI := fmt.Sprintf("%s/terminals?cols=%d&rows=%d&version=%s&shell=%s", consoleURI, cols, rows, terminalVersion, shell)
	termReq, err := http.NewRequest(http.MethodPost, termURI, bytes.NewBufferString("{}"))
	if err != nil {
		return nil, err
	}

	arm.SetCommonHeaders(termReq, token)
	arm.SetContentTypeJSON(termReq)

	resp, data, err := arm.ExecuteRequest(termReq)
	if err != nil {
		return nil, err
	}

	if err := arm.CheckStatus(resp.StatusCode); err != nil {
		return nil, fmt.Errorf("create terminal: %s, response: %s", resp.Status, string(data))
	}

	var terminal TerminalResponse
	if err := json.Unmarshal(data, &terminal); err != nil {
		return nil, err
	}

	return &terminal, nil
}

func DeleteConsole(token string) error {
	req, err := http.NewRequest(http.MethodDelete, consoleURL, bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	arm.SetCommonHeaders(req, token)
	req.Header.Set("content-type", "text/plain;charset=UTF-8")

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := arm.CheckStatus(resp.StatusCode, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete console: %s, response: %s", resp.Status, string(data))
	}

	return nil
}

func ResizeTerminal(token, consoleURI, terminalID string, cols, rows int) error {
	uri := fmt.Sprintf("%s/terminals/%s/size?cols=%d&rows=%d&version=%s", consoleURI, terminalID, cols, rows, terminalVersion)
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	arm.SetCommonHeaders(req, token)
	arm.SetContentTypeJSON(req)

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if err := arm.CheckStatus(resp.StatusCode); err != nil {
		return fmt.Errorf("resize terminal: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
