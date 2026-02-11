# 📦 demod

> **De**clarative git **mod**ule synchronizer

A tool that selectively syncs specific directories from multiple Git repositories.
Manage everything declaratively with a single TOML file.

## ✨ Features

- 🎯 **Sparse checkout** — Fetch only the paths you need, not the entire repo
- 📝 **Declarative config** — Manage all module dependencies in a single TOML file
- 🚀 **Concurrent** — Clone and sync multiple modules in parallel
- 🔍 **Dry-run** — Preview changes before applying them
- 🚫 **Exclude patterns** — Filter out unwanted files with glob patterns

## 📥 Installation

### Download binary

Download the latest binary from [GitHub Releases](https://github.com/mashiro/demod/releases).

### Go install

```bash
go install github.com/mashiro/demod/cmd/demod@latest
```

## 🚀 Usage

### Create a config file

Place a `demod.toml` in your project root.

```toml
dest_root = "vendor"

[[modules]]
name = "googleapis"
repo = "https://github.com/googleapis/googleapis.git"
revision = "master"
dest = "googleapis"
paths = [
  { src = "google/api", exclude = ["**/BUILD.bazel"] },
  { src = "google/rpc", exclude = ["**/BUILD.bazel"] },
]

[[modules]]
name = "github-rest-api"
repo = "https://github.com/github/rest-api-description.git"
revision = "main"
dest = "github"
paths = [
  { src = "descriptions-next/api.github.com", as = "openapi", exclude = ["**/*.yaml"] },
]
```

### Run sync

```bash
# Sync modules
demod sync

# Preview changes without writing
demod sync --dry-run
```

## ⚙️ Config Reference

### Top-level

| Key | Required | Description |
|-----|:--------:|-------------|
| `version` | | Config format version (default: `1`) |
| `dest_root` | | Root destination path for all modules |

### `[[modules]]`

| Key | Required | Description |
|-----|:--------:|-------------|
| `name` | ✅ | Module name |
| `repo` | ✅ | Git repository URL |
| `revision` | ✅ | Branch, tag, or commit hash |
| `dest` | ✅ | Destination directory |
| `paths` | ✅ | Array of paths to sync |

### `paths`

| Key | Required | Description |
|-----|:--------:|-------------|
| `src` | ✅ | Path within the source repository |
| `as` | | Destination directory name (defaults to `src`) |
| `exclude` | | Array of glob patterns to exclude |

## 🖥️ CLI Options

```
demod [flags] <command> [command-flags]
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--config, -c` | Config file path (default: `demod.toml`) |
| `--format, -f` | Log format (`text` / `json`) |
| `--no-color` | Disable colored output |
| `--verbose, -v` | Enable debug logging |

### Commands

| Command | Description |
|---------|-------------|
| `sync` | Sync modules (supports `--dry-run`) |
| `version` | Show version |

## 🛠️ Development

Toolchain is managed with [mise](https://mise.jdx.dev/).

```bash
# Build
mise run build

# Test
mise run test

# Lint
mise run lint

# Validate GoReleaser config
mise run release-check

# Local snapshot build
mise run release-snapshot
```
