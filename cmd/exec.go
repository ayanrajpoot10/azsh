package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/terminal"
)

var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Run a command in Cloud Shell",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExecCmd,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	settings, err := cloudshell.GetUserSettings(token)
	if err != nil {
		if cloudshell.IsUserSettingsNotFound(err) {
			return fmt.Errorf("Cloud Shell is not registered. Run 'azsh register' first")
		}
		return fmt.Errorf("user settings: %w", err)
	}

	consoleRes, err := cloudshell.ProvisionConsole(token, settings.PreferredOsType, settings.PreferredLocation)
	if err != nil {
		return fmt.Errorf("provision console: %w", err)
	}

	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, settings.PreferredShellType, 120, 30)
	if err != nil {
		return fmt.Errorf("negotiate terminal: %w", err)
	}

	wsURL, err := cloudshell.BuildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("build websocket URL: %w", err)
	}

	command := strings.Join(args, " ")
	return terminal.ExecCommand(wsURL, command)
}
