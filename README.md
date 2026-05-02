# azsh

A lightweight CLI tool to access Azure Cloud Shell directly from your terminal.

## Features

- **Azure Authentication** - Seamless OAuth2 authentication via device code flow
- **Cloud Shell Access** - Connect to Azure Cloud Shell without leaving your terminal
- **Terminal Emulation** - Full terminal support with resize handling
- **Token Caching** - Automatic token caching for faster subsequent connections

## Prerequisites

- Go (for building)
- Azure account with Cloud Shell enabled
- Linux, macOS, or WSL environment

## Installation

### From Source

```bash
go install github.com/ayanrajpoot10/azsh@latest
```

## Usage

```bash
azsh
```

First run will prompt for Azure authentication via device code flow. Subsequent runs will use cached credentials.

## How It Works

1. **Authentication** - Acquires OAuth2 token using Azure AD device code flow
2. **Settings** - Fetches your Cloud Shell preferences (OS type, location)
3. **Provisioning** - Creates a Cloud Shell console instance
4. **Terminal Setup** - Negotiates WebSocket connection for terminal I/O
5. **Connection** - Establishes interactive terminal session with proper TTY handling

## Authentication

- Uses Azure AD OAuth2 with device code flow
- Tokens are cached in `~/.azsh/token.json`
- Tenant information cached in `~/.azsh/tenant.json`
- Automatic token refresh for subsequent sessions

## License

MIT
