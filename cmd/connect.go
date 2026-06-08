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
		return fmt.Errorf("auth failed: %w", err)
	}

	settings, err := fetchUserSettings(token)
	if err != nil {
		return err
	}

	return openSession(token, settings)
}

func fetchUserSettings(token string) (*cloudshell.Properties, error) {
	fmt.Println("Fetching user settings...")
	settings, err := cloudshell.GetUserSettings(token)
	if err != nil {
		if cloudshell.IsUserSettingsNotFound(err) {
			return nil, fmt.Errorf("Cloud Shell is not registered. Run 'azsh register' first")
		}
		return nil, fmt.Errorf("user settings: %w", err)
	}
	return settings, nil
}

func openSession(token string, settings *cloudshell.Properties) error {
	location := settings.PreferredLocation
	osType := settings.PreferredOsType
	shellType := settings.PreferredShellType

	fmt.Print("Requesting a Cloud Shell... ")
	consoleRes, err := cloudshell.ProvisionConsole(token, osType, location)
	if err != nil {
		return fmt.Errorf("failed to provision console: %w", err)
	}
	fmt.Println("Succeeded.")

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120
		height = 30
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
