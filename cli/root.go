package cli

import (
	"fmt"
)

func Run(args []string) error {
	if len(args) == 0 {
		return connect()
	}

	command := args[0]

	switch command {
	case "logout":
		return logout()
	case "help", "-h", "--help":
		help()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\n\nUse 'azsh help' for usage information", command)
	}
}
