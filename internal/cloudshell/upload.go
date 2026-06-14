package cloudshell

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ayanrajpoot10/azsh/internal/arm"
)

func UploadFile(token, consoleURI, terminalID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("uploading-file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("create form: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	writer.Close()

	uri := fmt.Sprintf("%s/terminals/%s/upload", consoleURI, terminalID)
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	arm.SetCommonHeaders(req, token)
	req.Header.Set("content-type", writer.FormDataContentType())

	resp, data, err := arm.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload: %s, response: %s", resp.Status, string(data))
	}

	return nil
}
