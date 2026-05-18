package cloudshell

import (
	"fmt"
	"net/url"
)

const (
	wsScheme     = "wss"
	terminalPath = "/$hc/%s/terminals/%s"
)

func BuildWebSocketURL(consoleURI string, terminalID string) (string, error) {
	u, err := url.Parse(consoleURI)
	if err != nil {
		return "", err
	}

	u.Scheme = wsScheme
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	u.Path = fmt.Sprintf(terminalPath, path, terminalID)

	return u.String(), nil
}
