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
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	fmt.Println("Fetching subscriptions...")
	subscriptions, err := cloudshell.ListSubscriptions(token)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}
	if len(subscriptions) == 0 {
		return fmt.Errorf("no Azure subscriptions found")
	}

	subNames := make([]string, len(subscriptions))
	for i, s := range subscriptions {
		subNames[i] = fmt.Sprintf("%s (%s)", s.DisplayName, s.SubscriptionID)
	}
	subIdx, err := utils.PromptSelect("Select a subscription:", subNames)
	if err != nil {
		return err
	}
	sub := subscriptions[subIdx]

	fmt.Println("Fetching resource groups...")
	resourceGroups, err := cloudshell.ListResourceGroups(token, sub.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to list resource groups: %w", err)
	}
	if len(resourceGroups) == 0 {
		return fmt.Errorf("no resource groups found in '%s'. Create one in the Azure portal first", sub.DisplayName)
	}

	rgNames := make([]string, len(resourceGroups))
	for i, rg := range resourceGroups {
		rgNames[i] = fmt.Sprintf("%s (%s)", rg.Name, rg.Location)
	}
	rgIdx, err := utils.PromptSelect("Select a resource group:", rgNames)
	if err != nil {
		return err
	}
	rg := resourceGroups[rgIdx]

	fmt.Println("Registering Cloud Shell...")
	if err := cloudshell.RegisterUserSettings(token, sub.SubscriptionID, rg.Location); err != nil {
		return fmt.Errorf("failed to register Cloud Shell: %w", err)
	}
	fmt.Println("Cloud Shell registered successfully.")

	return nil
}
