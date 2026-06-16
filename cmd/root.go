package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "azsh",
	Short:         "A CLI client for Azure Cloud Shell",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runConnectCmd,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
