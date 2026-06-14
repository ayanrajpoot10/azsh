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

var shellConnect string

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVar(&shellConnect, "shell", "bash", "Shell type (bash or pwsh)")
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	settings, err := cloudshell.GetUserSettings(token)
	if err != nil {
		return fmt.Errorf("user settings: %w", err)
	}

	consoleRes, err := cloudshell.ProvisionConsole(token, settings.PreferredOsType, settings.PreferredLocation)
	if err != nil {
		return fmt.Errorf("provision console: %w", err)
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 120, 30
	}

	terminalInfo, err := cloudshell.CreateTerminal(token, consoleRes.Properties.URI, shellConnect, width, height)
	if err != nil {
		return fmt.Errorf("create terminal: %w", err)
	}

	wsURL, err := cloudshell.BuildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("build websocket URL: %w", err)
	}

	terminal.HandleResize(func(w, h int) {
		cloudshell.ResizeTerminal(token, consoleRes.Properties.URI, terminalInfo.ID, w, h)
	})

	if err := terminal.Connect(wsURL); err != nil {
		return fmt.Errorf("websocket: %w", err)
	}
	return nil
}
