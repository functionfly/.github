# `fly` — FunctionFly Developer CLI

The `fly` CLI is the primary developer interface for FunctionFly.

## Quick Start

```bash
fly login                    # Authenticate
fly init my-function         # Scaffold a new function
cd my-function
fly dev                      # Run locally at http://localhost:8787
fly publish                  # Publish to the global registry
fly test                     # Test the deployed function
```

## Commands

| Command | Description |
|---------|-------------|
| `fly login` | OAuth login (GitHub or Google) |
| `fly whoami` | Show current user |
| `fly logout` | Clear credentials |
| `fly init <name>` | Scaffold a new function project |
| `fly dev` | Run locally with hot reload |
| `fly publish` | Publish to registry |
| `fly publish --build` | Build then publish |
| `fly test` | Test deployed function |
| `fly update patch` | Bump version (patch/minor/major) |
| `fly stats` | View usage statistics |
| `fly logs` | View recent logs |
| `fly logs --follow` | Stream live logs |
| `fly rollback` | Roll back to previous version |
| `fly env list/set/get/unset` | Manage environment variables |
| `fly secrets list/set/unset` | Manage secrets |
| `fly completion bash/zsh/fish` | Shell completion |

## JSON Output

All commands support `--json` for CI/CD:

```bash
fly publish --json
fly stats --json
fly whoami --json
```

## Config

Global config: `~/.functionfly/config.yaml`
Credentials: `~/.functionfly/credentials.json`
