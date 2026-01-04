# Binarius - Universal Binary Version Manager

Binarius is a universal binary version manager that provides version switching for any single-binary CLI tool.

## Currently Supported Tools

- **`terraform`** - HashiCorp Terraform
- **`tofu`** - OpenTofu (open-source Terraform alternative)
- **`terragrunt`** - Terragrunt (Terraform wrapper)

Adding support for new tools is straightforward - see the extensibility documentation.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform:

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 (x86_64) | [binarius_linux_amd64.tar.gz](https://github.com/NIXKnight/Binarius/releases/latest/download/binarius_linux_amd64.tar.gz) |
| Linux | arm64 (aarch64) | [binarius_linux_arm64.tar.gz](https://github.com/NIXKnight/Binarius/releases/latest/download/binarius_linux_arm64.tar.gz) |

### Quick Install (Linux)

One-liner installation for Linux amd64:

```bash
curl -sL https://github.com/NIXKnight/Binarius/releases/latest/download/binarius_linux_amd64.tar.gz | tar xz -C ~/.local/bin binarius
```

One-liner installation for Linux arm64:

```bash
curl -sL https://github.com/NIXKnight/Binarius/releases/latest/download/binarius_linux_arm64.tar.gz | tar xz -C ~/.local/bin binarius
```

### Verify Checksum

For security, verify the downloaded archive against the published checksums:

```bash
# Download the checksums file
curl -sLO https://github.com/NIXKnight/Binarius/releases/latest/download/checksums.txt

# Verify the archive (example for amd64)
sha256sum -c checksums.txt --ignore-missing
```

### Build from Source

If you prefer to build from source:

```bash
git clone https://github.com/NIXKnight/Binarius.git
cd Binarius
make build
mv binarius ~/.local/bin/
# Or for system-wide install:
sudo mv binarius /usr/local/bin/
```

### PATH Configuration

Ensure `~/.local/bin` is in your PATH:

```bash
# Check if already in PATH
echo $PATH | grep -q "$HOME/.local/bin" && echo "Already in PATH" || echo "Not in PATH"

# Add to PATH (if needed)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Quick Start

```bash
# Initialize Binarius (creates ~/.binarius/ directory structure)
binarius init

# Install a tool version
binarius install terraform@v1.6.0

# Set active version (creates symlink)
binarius use terraform@v1.6.0

# Use the tool directly
terraform version

# List installed versions
binarius list

# View detailed tool information
binarius info terraform

# Uninstall a version
binarius uninstall terraform@v1.6.0
```

## Directory Structure

```
~/.binarius/                          # Binarius home
├── config.yaml                       # Global configuration
├── installation.json                 # Installation registry
├── tools/                            # Installed binaries
│   └── <tool>/
│       └── <version>/
│           └── <binary>
└── cache/                            # Downloaded archives

~/.local/bin/                         # Symlinks (in PATH)
└── <tool> → ~/.binarius/tools/<tool>/<version>/<binary>
```

## Usage Examples

### Installing Versions

```bash
# Install latest version
binarius install terraform@latest

# Install specific version
binarius install terraform@v1.5.0

# Install multiple versions
binarius install terraform@v1.5.0 terraform@v1.6.0

# Install multiple tools
binarius install terraform@latest tofu@latest terragrunt@latest
```

### Switching Versions

```bash
# Switch to specific version
binarius use terraform@v1.6.0

# Switch back to previous version
binarius use terraform@v1.5.0

# Verify active version
terraform version
binarius info terraform
```

### Managing Installations

```bash
# List all installed tools and versions
binarius list

# List versions for specific tool
binarius list terraform

# View detailed information
binarius info terraform

# Uninstall specific version
binarius uninstall terraform@v1.5.0

# Uninstall all versions of a tool
binarius uninstall terraform
```

## Configuration

### Global Defaults

Edit `~/.binarius/config.yaml` to set default versions:

```yaml
defaults:
  terraform: v1.6.0
  tofu: v1.6.0
  terragrunt: v0.54.0

paths:
  binarius_home: ~/.binarius
  bin_dir: ~/.local/bin
  cache_dir: ~/.binarius/cache
```

### Installation Registry

Binarius maintains a registry of installed versions in `~/.binarius/installation.json`:

```json
{
  "terraform": {
    "v1.6.0": {
      "installed_at": "2024-01-15T10:30:00Z",
      "binary_path": "~/.binarius/tools/terraform/v1.6.0/terraform",
      "size_bytes": 25678901,
      "source_url": "https://releases.hashicorp.com/terraform/1.6.0/...",
      "checksum": "abc123..."
    }
  }
}
```

## Extensibility

Binarius is designed to support any single-binary CLI tool. While the initial focus is infrastructure tooling (terraform, opentofu, terragrunt), the architecture supports extending to any CLI tool with downloadable binaries.

See project documentation for details on adding new tool support.

## License

MIT. See LICENSE file.
