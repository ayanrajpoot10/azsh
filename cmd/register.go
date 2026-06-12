package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/utils"
)

const diskSizeGB = 5

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
		return fmt.Errorf("authenticate: %w", err)
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

	subID := subscriptions[subIdx].SubscriptionID

	csType, err := utils.PromptSelect("Cloud Shell type:", []string{
		"Ephemeral (no mounted storage)",
		"Mounted storage (with file share)",
	})
	if err != nil {
		return err
	}

	var location string
	var storageProfile *cloudshell.StorageProfile

	switch csType {
	case 0:
		location, err = utils.PromptInput("Enter preferred Azure region (e.g. centralindia):")
		if err != nil {
			return err
		}
		if location == "" {
			return fmt.Errorf("region is required")
		}
		storageProfile = nil

	case 1:
		storageIdx, err := utils.PromptSelect("Storage setup:", []string{
			"Select existing storage account",
			"Auto-setup (create resource group + storage account)",
			"Enter storage details manually",
		})
		if err != nil {
			return err
		}

		switch storageIdx {
		case 0:
			storageProfile, location, err = selectExistingStorage(token, subID)
		case 1:
			location, err = utils.PromptInput("Enter Azure region for new resources (e.g. centralindia):")
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("region is required")
			}
			storageProfile, err = autoSetupStorage(token, subID, location)
		case 2:
			location, err = utils.PromptInput("Enter Azure region (e.g. centralindia):")
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("region is required")
			}
			storageProfile, err = manualStorage()
		}
		if err != nil {
			return err
		}
	}

	fmt.Println("Registering Cloud Shell...")
	if err := cloudshell.RegisterUserSettings(token, subID, location, storageProfile); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	fmt.Println("Cloud Shell registered.")
	return nil
}

func selectExistingStorage(token, subID string) (*cloudshell.StorageProfile, string, error) {
	fmt.Println("Fetching resource groups...")
	rgs, err := cloudshell.ListResourceGroups(token, subID)
	if err != nil {
		return nil, "", err
	}
	if len(rgs) == 0 {
		return nil, "", fmt.Errorf("no resource groups in subscription")
	}

	rgNames := make([]string, len(rgs))
	for i, rg := range rgs {
		rgNames[i] = fmt.Sprintf("%s (%s)", rg.Name, rg.Location)
	}
	rgIdx, err := utils.PromptSelect("Select a resource group:", rgNames)
	if err != nil {
		return nil, "", err
	}

	rgName := rgs[rgIdx].Name
	location := rgs[rgIdx].Location

	fmt.Println("Fetching storage accounts...")
	accounts, err := cloudshell.ListStorageAccounts(token, subID, rgName)
	if err != nil {
		return nil, "", fmt.Errorf("list storage accounts: %w", err)
	}
	if len(accounts) == 0 {
		return nil, "", fmt.Errorf("no storage accounts found in resource group %s", rgName)
	}

	acctNames := make([]string, len(accounts))
	for i, a := range accounts {
		acctNames[i] = fmt.Sprintf("%s (%s)", a.Name, a.Location)
	}
	acctIdx, err := utils.PromptSelect("Select a storage account:", acctNames)
	if err != nil {
		return nil, "", err
	}

	_, fileShareName := cloudshell.GenerateStorageNames(token)

	return &cloudshell.StorageProfile{
		StorageAccountResourceID: accounts[acctIdx].ID,
		FileShareName:            fileShareName,
		DiskSizeInGB:             diskSizeGB,
	}, location, nil
}

func autoSetupStorage(token, subID, location string) (*cloudshell.StorageProfile, error) {
	acctName, fileShareName := cloudshell.GenerateStorageNames(token)
	if acctName == "" {
		return nil, fmt.Errorf("could not generate storage account name from token")
	}

	autoRGName := "cloud-shell-storage-" + location

	fmt.Printf("Creating resource group %s...\n", autoRGName)
	if err := cloudshell.CreateResourceGroup(token, subID, autoRGName, location); err != nil {
		return nil, fmt.Errorf("create resource group: %w", err)
	}
	fmt.Println("Resource group created.")

	fmt.Printf("Creating storage account %s...\n", acctName)
	if err := cloudshell.CreateStorageAccount(token, subID, autoRGName, acctName, location); err != nil {
		return nil, fmt.Errorf("create storage account: %w", err)
	}
	fmt.Println("Storage account created.")

	fmt.Println("Registering Microsoft.CloudShell resource provider...")
	if err := cloudshell.RegisterCloudShellRP(token, subID); err != nil {
		return nil, fmt.Errorf("register RP: %w", err)
	}

	storageAccountID := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s",
		subID, autoRGName, acctName,
	)

	return &cloudshell.StorageProfile{
		StorageAccountResourceID: storageAccountID,
		FileShareName:            fileShareName,
		DiskSizeInGB:             diskSizeGB,
	}, nil
}

func manualStorage() (*cloudshell.StorageProfile, error) {
	resourceID, err := utils.PromptInput("Enter storage account resource ID:")
	if err != nil {
		return nil, err
	}
	if resourceID == "" {
		return nil, fmt.Errorf("storage account resource ID is required")
	}

	fileShareName, err := utils.PromptInput("Enter file share name:")
	if err != nil {
		return nil, err
	}
	if fileShareName == "" {
		return nil, fmt.Errorf("file share name is required")
	}

	return &cloudshell.StorageProfile{
		StorageAccountResourceID: resourceID,
		FileShareName:            fileShareName,
		DiskSizeInGB:             diskSizeGB,
	}, nil
}
