package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/ayanrajpoot10/azsh/internal/auth"
	"github.com/ayanrajpoot10/azsh/internal/cloudshell"
	"github.com/ayanrajpoot10/azsh/internal/terminal"
)

const (
	defaultShellType = "bash"
	defaultOSType    = "Linux"
	defaultWidth     = 120
	defaultHeight    = 30
	wsScheme         = "wss"
	terminalPath     = "/$hc/%s/terminals/%s"
)

type ConnectOptions struct {
	Shell      string
	Location   string
	ForceLogin bool
}

type CLI struct {
	fs *pflag.FlagSet
}

func NewCLI() *CLI {
	return &CLI{
		fs: pflag.NewFlagSet("azsh", pflag.ContinueOnError),
	}
}

func (c *CLI) handleConnect(args []string) error {
	opts := &ConnectOptions{
		Shell: defaultShellType,
	}

	fs := pflag.NewFlagSet("connect", pflag.ContinueOnError)
	fs.StringVar(&opts.Shell, "shell", defaultShellType, "Shell type (bash or zsh)")
	fs.StringVar(&opts.Location, "location", "", "Preferred location for Cloud Shell")
	fs.BoolVar(&opts.ForceLogin, "force-login", false, "Force login even if credentials are cached")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	return connectCloudShell(opts)
}

func (c *CLI) handleLogout() error {
	return auth.Logout()
}

func (c *CLI) handleHelp() {
	help := `Usage: azsh [command] [flags]

Commands:
  connect              Connect to Azure Cloud Shell (default)
  logout               Logout and clear cached credentials
  help                 Show this help message

Connect Flags:
  --shell string       Shell type to use: bash or zsh (default: bash)
  --location string    Preferred location for Cloud Shell
  --force-login        Force login even if credentials are cached
  -h, --help          Show help message

Examples:
  azsh                                    # Connect with defaults
  azsh connect --shell zsh                # Connect with zsh shell
  azsh connect --location eastus          # Connect to specific location
  azsh connect --force-login              # Force login prompt
  azsh logout                             # Logout and clear cache
`
	fmt.Print(help)
}

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return c.handleConnect([]string{})
	}

	command := args[0]

	switch command {
	case "connect":
		return c.handleConnect(args[1:])
	case "logout":
		return c.handleLogout()
	case "help", "-h", "--help":
		c.handleHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\n\nUse 'azsh help' for usage information", command)
	}
}

func buildWebSocketURL(consoleURI string, terminalID string) (string, error) {
	u, err := url.Parse(consoleURI)
	if err != nil {
		return "", err
	}

	u.Scheme = wsScheme
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	u.Path = fmt.Sprintf(terminalPath, path, terminalID)

	return u.String(), nil
}

func handleWindowResize(token, consoleURI, terminalID string) {
	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)

	go func() {
		for range sigWinch {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				cloudshell.ResizeTerminal(token, consoleURI, terminalID, w, h)
			}
		}
	}()
}

func connectCloudShell(opts *ConnectOptions) error {
	log.Println("Authenticating...")
	token, err := auth.GetToken(opts.ForceLogin)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	log.Println("Fetching user settings...")
	settings, err := cloudshell.UserSettings(token)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	location := opts.Location
	if location == "" {
		location = settings.PreferredLocation
	}

	log.Print("Requesting a Cloud Shell. ")
	consoleRes, err := cloudshell.ProvisionConsole(token, defaultOSType, location)
	if err != nil {
		return fmt.Errorf("failed to provision console: %w", err)
	}
	log.Println("Succeeded.")

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultWidth
		height = defaultHeight
	}

	log.Println("Connecting terminal...")
	terminalInfo, err := cloudshell.NegotiateTerminal(token, consoleRes.Properties.URI, opts.Shell, width, height)
	if err != nil {
		return fmt.Errorf("failed to negotiate terminal: %w", err)
	}

	wsURL, err := buildWebSocketURL(consoleRes.Properties.URI, terminalInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to build websocket URL: %w", err)
	}

	handleWindowResize(token, consoleRes.Properties.URI, terminalInfo.ID)

	if err := terminal.Connect(wsURL); err != nil {
		return fmt.Errorf("websocket connection error: %w", err)
	}

	return nil
}

