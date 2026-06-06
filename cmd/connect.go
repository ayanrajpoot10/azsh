package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/terminal"
)

const (
	defaultShellType = "bash"
	defaultOSType    = "Linux"
	defaultWidth     = 120
	defaultHeight    = 30
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Azure Cloud Shell",
	RunE:  runConnectCmd,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	settings, err := getUserSettings(token)
	if err != nil {
		return err
	}

	return startSession(token, settings)
}

func getUserSettings(token string) (*cloudshell.Properties, error) {
	fmt.Println("Fetching user settings...")
	settings, err := cloudshell.GetUserSettings(token)
	if err != nil {
		if cloudshell.IsUserSettingsNotFound(err) {
			return nil, fmt.Errorf("Cloud Shell is not registered. Run 'azsh register' first")
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return settings, nil
}

func startSession(token string, settings *cloudshell.Properties) error {
	location := settings.PreferredLocation
	osType := settings.PreferredOsType
	if osType == "" {
		osType = defaultOSType
	}
	shellType := settings.PreferredShellType
	if shellType == "" {
		shellType = defaultShellType
	}

	fmt.Print("Requesting a Cloud Shell. ")
	consoleRes, err := cloudshell.ProvisionConsole(token, osType, location)
	if err != nil {
		return fmt.Errorf("failed to provision console: %w", err)
	}
	fmt.Println("Succeeded.")

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultWidth
		height = defaultHeight
	}

	fmt.Println("Connecting terminal...")
	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, shellType, width, height)
	if err != nil {
		return fmt.Errorf("failed to negotiate terminal: %w", err)
	}

	wsURL, err := cloudshell.BuildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to build websocket URL: %w", err)
	}

	terminal.HandleResize(func(w, h int) {
		cloudshell.ResizeTerminal(token, consoleRes.Properties.URI, terminalInfo.ID, w, h)
	})

	if err := terminal.Connect(wsURL); err != nil {
		return fmt.Errorf("websocket connection error: %w", err)
	}

	return nil
}
