package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file to Cloud Shell",
	Args:  cobra.ExactArgs(1),
	RunE:  runUploadCmd,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func runUploadCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
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

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120
		height = 30
	}

	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, settings.PreferredShellType, width, height)
	if err != nil {
		return fmt.Errorf("negotiate terminal: %w", err)
	}

	filePath := args[0]
	fmt.Printf("Uploading %s...\n", filePath)
	if err := cloudshell.UploadFile(token, consoleRes.Properties.URI, terminalInfo.ID, filePath); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	fmt.Println("Upload complete.")
	return nil
}
