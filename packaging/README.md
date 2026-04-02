# FunctionFly CLI packaging

## Install script (Linux / macOS)

From the repo or directly:

```bash
curl -fsSL https://raw.githubusercontent.com/functionfly/functionfly/main/scripts/install.sh | bash
```

Options:

- `VERSION=v1.0.0` — install a specific version (default: latest)
- `BINDIR=~/.local/bin` — install to this directory (default: `~/.local/bin` if writable, else `/usr/local/bin`)

## Homebrew (macOS / Linux)

When a tap is set up (e.g. `homebrew-functionfly`):

```bash
brew tap functionfly/tap
brew install ffly
```

To maintain the formula: copy `homebrew/ffly.rb` into your tap and update `version` and each `sha256` from the [releases](https://github.com/functionfly/fly/releases) checksums file for the matching asset.

## Windows

- **Scoop**: After a release, add a manifest that points at the Windows zip from [releases](https://github.com/functionfly/functionfly/releases) (e.g. `fly_<version>_windows_x86_64.zip`). Then: `scoop install fly`.
- **Chocolatey**: Similarly, add a package that downloads the Windows zip and places `fly.exe` in the path. Then: `choco install fly`.

Binary artifact names follow GoReleaser: `fly_<version>_windows_x86_64.zip` (no arm64 for Windows in current config).

## Upgrading

- **Install script**: Re-run with `VERSION=latest` (or a specific version).
- **Homebrew**: `brew upgrade ffly` (after the tap is updated).
- **Scoop**: `scoop update fly`.
- **Chocolatey**: `choco upgrade fly`.
