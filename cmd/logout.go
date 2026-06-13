package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/utils"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out of Azure",
	RunE:  runLogoutCmd,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogoutCmd(cmd *cobra.Command, args []string) error {
	path, err := utils.CachePath("token.json")
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not logged in")
		}
		return fmt.Errorf("remove token.json: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
}
