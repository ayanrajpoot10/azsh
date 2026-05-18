package cli

import (
	"fmt"
)

func handleHelp() {
	help := `Usage: azsh [command]

Commands:
  logout               Logout and clear cached credentials

Examples:
  azsh                 # Connect with defaults
  azsh logout          # Logout and clear cache
`
	fmt.Print(help)
}
