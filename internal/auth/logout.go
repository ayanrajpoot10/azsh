package auth

import (
	"fmt"
	"os"
)

func Logout() error {
	path, err := getFilePath(tokenFile)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not logged in: no cached credentials found")
		}
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
}
