package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "azsh",
	Short: "Lightweight CLI for Azure Cloud Shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		return connectCmd.RunE(connectCmd, args)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
