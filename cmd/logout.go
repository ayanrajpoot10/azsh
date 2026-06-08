package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cached credentials",
	RunE:  runLogoutCmd,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogoutCmd(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".azsh", "token.json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not logged in")
		}
		return fmt.Errorf("remove token.json: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
}
