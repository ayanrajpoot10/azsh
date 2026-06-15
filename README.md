# azsh

`azsh` is a CLI client for [Azure Cloud Shell](https://learn.microsoft.com/en-us/azure/cloud-shell/overview) that lets you access a fully preconfigured environment with tools like `az`, `kubectl`, and `docker` directly from your local terminal. No browser or Azure CLI installation needed.

> **Note:** This tool is for Linux and macOS. On Windows, Azure Cloud Shell is available natively through Windows Terminal.

![azsh demo](assets/azsh.gif)

---

## Features

- **Interactive Cloud Shell** — full terminal session with bash or PowerShell
- **One-shot commands** — run `az`, `kubectl`, or any command and get the output
- **File upload** — push local files into your Cloud Shell environment
- **Azure auth via MSAL** — OAuth2 device code flow with token caching
- **Settings & console caching** — stored locally in `~/.azsh/`
- **Automatic terminal resize** — SIGWINCH signals forwarded to the Cloud Shell terminal

---

## Prerequisites

- An Azure account with an **active subscription**
- The Cloud Shell resource provider may need to be registered — `azsh register` handles this

---

## Installation

### go install

```bash
go install github.com/ayanrajpoot10/azsh@latest
```

### Download a release

Download the latest archive from [GitHub Releases](https://github.com/ayanrajpoot10/azsh/releases).

### Arch Linux

Install from the AUR:

```bash
yay -S azsh-bin
```

### Build from source

```bash
git clone https://github.com/ayanrajpoot10/azsh.git
cd azsh
go build
```

---

## Quick Start

```bash
azsh
```

After the first run, subsequent sessions reuse cached tokens and settings.

---

## Command Reference

### `azsh` / `azsh connect`

Start an interactive Cloud Shell session.

```bash
azsh                        # default: bash
azsh connect --shell pwsh   # PowerShell
azsh connect --shell bash   # bash (explicit)
```

Once connected, type commands normally. Exit with `exit` or `Ctrl-D`.

### `azsh login`

Authenticate with Azure using OAuth2 device code flow. Opens your browser for sign-in.

```bash
azsh login
```

Tokens are cached to `~/.azsh/token.json`.

### `azsh logout`

Remove the cached authentication token.

```bash
azsh logout
```

### `azsh register`

One-time setup to configure Cloud Shell for your Azure account.

```bash
azsh register
```

Walks you through:
1. Subscription selection (if you have multiple)
2. Cloud Shell type — ephemeral (no storage) or with mounted storage
3. Storage configuration — existing storage account, auto-setup, or manual

### `azsh exec`

Run a command non-interactively and print the output.

```bash
azsh exec "az account show"
azsh exec "df -h"
```

Useful for scripting, automation, or one-off queries without entering an interactive session.

### `azsh upload`

Upload a local file to your Cloud Shell home directory.

```bash
azsh upload ./script.sh
azsh upload /path/to/config.yaml
```

The file lands in `$HOME/` in the Cloud Shell environment.

### `azsh reset`

Delete the Cloud Shell console, user settings, and clear all local caches.

```bash
azsh reset
```

Use this if you hit provisioning limits or want to start fresh.

---

## FAQ

<details>
<summary>Why this tool instead of installing the Azure CLI locally or using Cloud Shell in a browser?</summary>

I used the Azure CLI locally, but its installed size is around 800 MB. I prefer a clean system without bloated packages. Cloud Shell provides a preconfigured environment with tools like `az` already installed, so there's no need to install anything locally. However, using Cloud Shell in a browser is a poor user experience. `azsh` brings that environment to your local terminal.
</details>

---

## License

MIT — see [LICENSE](LICENSE).

Copyright (c) 2026 Ayan Rajpoot.
