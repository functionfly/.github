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
brew install ff
```

To maintain the formula: copy `homebrew/ff.rb` into your tap and update `version` and each `sha256` from the [releases](https://github.com/functionfly/fly/releases) checksums file for the matching asset.

## Windows

- **Scoop**: After a release, add a manifest that points at the Windows zip from [releases](https://github.com/functionfly/fly/releases) (e.g. `ff_<version>_windows_x86_64.zip`). Then: `scoop install ff`.
- **Chocolatey**: Similarly, add a package that downloads the Windows zip and places `ff.exe` in the path. Then: `choco install ff`.

Binary artifact names follow GoReleaser: `ff_<version>_windows_x86_64.zip` (no arm64 for Windows in current config).

## Upgrading

- **Install script**: Re-run with `VERSION=latest` (or a specific version).
- **Homebrew**: `brew upgrade ff` (after the tap is updated).
- **Scoop**: `scoop update ff`.
- **Chocolatey**: `choco upgrade ff`.
