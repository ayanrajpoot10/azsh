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
	t, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	settings, err := cloudshell.GetUserSettings(t)
	if err != nil {
		if cloudshell.IsUserSettingsNotFound(err) {
			return fmt.Errorf("Cloud Shell is not registered. Run 'azsh register' first")
		}
		return fmt.Errorf("user settings: %w", err)
	}

	consoleRes, err := cloudshell.ProvisionConsole(t, settings.PreferredOsType, settings.PreferredLocation)
	if err != nil {
		return fmt.Errorf("provision console: %w", err)
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120
		height = 30
	}

	terminalInfo, err := cloudshell.NegotiateTerminal(t, consoleRes.Properties.URI, settings.PreferredShellType, width, height)
	if err != nil {
		return fmt.Errorf("negotiate terminal: %w", err)
	}

	wsURL, err := cloudshell.BuildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("build websocket URL: %w", err)
	}

	terminal.HandleResize(func(w, h int) {
		cloudshell.ResizeTerminal(t, consoleRes.Properties.URI, terminalInfo.ID, w, h)
	})

	if err := terminal.Connect(wsURL); err != nil {
		return fmt.Errorf("websocket: %w", err)
	}
	return nil
}
