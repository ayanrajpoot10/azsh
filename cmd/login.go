package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Azure",
	RunE:  runLoginCmd,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLoginCmd(cmd *cobra.Command, args []string) error {
	if _, err := auth.Authenticate(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	fmt.Println("Login successful.")
	return nil
}
