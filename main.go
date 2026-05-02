package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"azsh/internal/auth"
	"azsh/internal/cloudshell"
	"azsh/internal/terminal"
)

func main() {
	token, err := auth.GetToken()
	if err != nil {
		log.Fatalf("failed to get auth token: %v", err)
	}

	settings, err := cloudshell.UserSettings(token)
	if err != nil {
		log.Fatalf("failed to get user settings: %v", err)
	}

	log.Print("Requesting a Cloud Shell. ")
	consoleRes, err := cloudshell.ProvisionConsole(token, settings.PreferredOsType, settings.PreferredLocation)
	if err != nil {
		log.Fatalf("failed to provision console: %v", err)
	}
	log.Println("Succeeded.")

	shellType := "bash"

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120
		height = 30
	}

	log.Println("Connecting terminal...")

	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, shellType, width, height)
	if err != nil {
		log.Fatalf("failed to negotiate terminal: %v", err)
	}

	u, err := url.Parse(consoleRes.Properties.URI)
	if err != nil {
		log.Fatalf("failed to parse console URI: %v", err)
	}

	u.Scheme = "wss"
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	u.Path = fmt.Sprintf("/$hc/%s/terminals/%s", path, terminalInfo.ID)

	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	go func() {
		for range sigWinch {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				cloudshell.ResizeTerminal(token, consoleRes.Properties.URI, terminalInfo.ID, w, h)
			}
		}
	}()

	err = terminal.Connect(u.String())
	if err != nil {
		log.Fatalf("WebSocket connection error: %v", err)
	}
}
