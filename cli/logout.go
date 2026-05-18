package cli

import (
	"github.com/ayanrajpoot10/azsh/internal/auth"
)

func handleLogout() error {
	return auth.Logout()
}
