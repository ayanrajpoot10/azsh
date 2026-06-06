package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "azsh",
	Short: "Connect to Azure Cloud Shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		return connectCmd.RunE(connectCmd, args)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
