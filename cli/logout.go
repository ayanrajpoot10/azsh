package cli

import (
	"github.com/ayanrajpoot10/azsh/internal/auth"
)

func logout() error {
	return auth.Logout()
}
