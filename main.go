package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/terminal"
)

const (
	defaultShellType = "bash"
	defaultWidth     = 120
	defaultHeight    = 30
	wsScheme         = "wss"
	terminalPath     = "/$hc/%s/terminals/%s"
)

func buildWebSocketURL(consoleURI string, terminalID string) (string, error) {
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

func handleWindowResize(token, consoleURI, terminalID string) {
	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)

	go func() {
		for range sigWinch {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				cloudshell.ResizeTerminal(token, consoleURI, terminalID, w, h)
			}
		}
	}()
}

func run() error {
	log.Println("Authenticating...")
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	log.Println("Fetching user settings...")
	settings, err := cloudshell.UserSettings(token)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	log.Print("Requesting a Cloud Shell. ")
	consoleRes, err := cloudshell.ProvisionConsole(token, settings.PreferredOsType, settings.PreferredLocation)
	if err != nil {
		return fmt.Errorf("failed to provision console: %w", err)
	}
	log.Println("Succeeded.")

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultWidth
		height = defaultHeight
	}

	log.Println("Connecting terminal...")
	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, defaultShellType, width, height)
	if err != nil {
		return fmt.Errorf("failed to negotiate terminal: %w", err)
	}

	wsURL, err := buildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to build websocket URL: %w", err)
	}

	handleWindowResize(token, consoleRes.Properties.URI, terminalInfo.ID)

	if err := terminal.Connect(wsURL); err != nil {
		return fmt.Errorf("websocket connection error: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}
