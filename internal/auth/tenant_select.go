package auth

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func selectTenant(tenants []Tenant) (string, error) {
	fmt.Println("\nMultiple tenants found. Please select one:")
	fmt.Println()

	for i, tenant := range tenants {
		fmt.Printf("  [%d] %s\n", i+1, tenant.ID)
	}

	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter your choice (1-" + strconv.Itoa(len(tenants)) + "): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(tenants) {
			fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(tenants))
			continue
		}

		return tenants[choice-1].ID, nil
	}
}
