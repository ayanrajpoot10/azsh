package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Cloud Shell console and settings",
	RunE:  runResetCmd,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runResetCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	fmt.Println("Deleting Cloud Shell console...")
	if err := cloudshell.DeleteConsole(token); err != nil {
		return fmt.Errorf("delete console: %w", err)
	}
	fmt.Println("Console deleted.")

	fmt.Println("Deleting Cloud Shell user settings...")
	if err := cloudshell.DeleteUserSettings(token); err != nil {
		return fmt.Errorf("delete user settings: %w", err)
	}
	fmt.Println("User settings deleted.")

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".azsh")
	for _, name := range []string{"settings.json", "console.json"} {
		path := filepath.Join(cacheDir, name)
		if err := os.Remove(path); err == nil {
			fmt.Printf("Removed local cache: %s\n", name)
		}
	}

	fmt.Println("Cloud Shell reset complete. Run 'azsh register' to set up again.")
	return nil
}
