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

	azshDir := filepath.Join(home, ".azsh")

	removed := false
	for _, name := range []string{"token.json", "tenant.json"} {
		path := filepath.Join(azshDir, name)
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove %s: %w", name, err)
			}
			continue
		}
		removed = true
	}

	if !removed {
		return fmt.Errorf("not logged in: no cached credentials found")
	}

	fmt.Println("Logged out successfully.")
	return nil
}
