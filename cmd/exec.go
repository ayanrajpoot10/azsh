package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/terminal"
)

var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Run a command on Cloud Shell and print its output",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExecCmd,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExecCmd(cmd *cobra.Command, args []string) error {
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
		width, height = 120, 30
	}

	terminalInfo, err := cloudshell.NegotiateTerminal(t, consoleRes.Properties.URI, settings.PreferredShellType, width, height)
	if err != nil {
		return fmt.Errorf("negotiate terminal: %w", err)
	}

	wsURL, err := cloudshell.BuildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("build websocket URL: %w", err)
	}

	command := strings.Join(args, " ")
	err = terminal.ExecCommand(wsURL, command)
	return err
}
