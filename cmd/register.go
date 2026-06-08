package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/utils"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register Cloud Shell for your Azure account",
	RunE:  runRegisterCmd,
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

func runRegisterCmd(cmd *cobra.Command, args []string) error {
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	fmt.Println("Fetching subscriptions...")
	subscriptions, err := cloudshell.ListSubscriptions(token)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}
	if len(subscriptions) == 0 {
		return fmt.Errorf("no subscriptions found")
	}

	subNames := make([]string, len(subscriptions))
	for i, s := range subscriptions {
		subNames[i] = fmt.Sprintf("%s (%s)", s.DisplayName, s.SubscriptionID)
	}
	subIdx, err := utils.PromptSelect("Select a subscription:", subNames)
	if err != nil {
		return err
	}

	fmt.Println("Fetching resource groups...")
	rgs, err := cloudshell.ListResourceGroups(token, subscriptions[subIdx].SubscriptionID)
	if err != nil {
		return err
	}
	if len(rgs) == 0 {
		return fmt.Errorf("no resource groups in subscription")
	}

	rgNames := make([]string, len(rgs))
	for i, rg := range rgs {
		rgNames[i] = fmt.Sprintf("%s (%s)", rg.Name, rg.Location)
	}
	rgIdx, err := utils.PromptSelect("Select a resource group:", rgNames)
	if err != nil {
		return err
	}

	fmt.Println("Registering Cloud Shell...")
	if err := cloudshell.RegisterUserSettings(token, subscriptions[subIdx].SubscriptionID, rgs[rgIdx].Location); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	fmt.Println("Cloud Shell registered.")
	return nil
}
