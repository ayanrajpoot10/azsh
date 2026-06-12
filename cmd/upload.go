package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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

	terminalInfo, err := cloudshell.NegotiateTerminal(t, consoleRes.Properties.URI, settings.PreferredShellType, 120, 30)
	if err != nil {
		return fmt.Errorf("negotiate terminal: %w", err)
	}

	filePath := args[0]
	fmt.Printf("Uploading %s...\n", filePath)
	if err := cloudshell.UploadFile(t, consoleRes.Properties.URI, terminalInfo.ID, filePath); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	fmt.Println("Upload complete.")
	return nil
}
