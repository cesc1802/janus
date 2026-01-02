# Deployment & Installation Guide

## Installation Scripts

**Phase 01-02 Completion:** Automated installation scripts for Unix-like systems (Linux, macOS) and Windows.

### Unix Installation Script

**Completion:** Automated installation script for Unix-like systems (Linux, macOS).

### Overview

The `scripts/install.sh` script provides a robust, cross-platform method to download and install `migrate-tool` from GitHub releases. It includes:
- Automatic OS/architecture detection (Linux/macOS, amd64/arm64)
- SHA256 checksum verification
- Dependency checking (curl/wget, tar, sha256sum/shasum)
- Automatic fallback between curl and wget
- Colored output with NO_COLOR support
- Installation to custom directories
- Version-specific or latest version installation

### Usage

#### Basic Installation (Latest Version)

```bash
./scripts/install.sh
```

Installs latest `migrate-tool` release to `/usr/local/bin`.

#### System-Wide Installation (Requires sudo)

```bash
sudo ./scripts/install.sh
```

#### User-Local Installation (No sudo required)

```bash
INSTALL_DIR=~/.local/bin ./scripts/install.sh
```

Ensure `~/.local/bin` is in your PATH:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

#### Install Specific Version

```bash
./scripts/install.sh --version v0.0.3
```

#### Custom Installation Directory

```bash
./scripts/install.sh --dir /opt/janus
```

### Options

| Option | Description | Example |
|--------|-------------|---------|
| `-v, --version VERSION` | Install specific version | `./install.sh --version v0.0.3` |
| `-d, --dir DIRECTORY` | Custom install directory | `./install.sh --dir ~/.local/bin` |
| `-h, --help` | Show help message | `./install.sh --help` |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `INSTALL_DIR` | Installation directory | `/usr/local/bin` |
| `NO_COLOR` | Disable colored output | (not set = colored) |

### Examples

```bash
# Install latest to system directory
sudo ./scripts/install.sh

# Install specific version to home directory
INSTALL_DIR=~/.local/bin ./scripts/install.sh --version v0.0.3

# Install with custom directory via flag
./scripts/install.sh -d /opt/tools

# Install with NO_COLOR output
NO_COLOR=1 ./scripts/install.sh
```

### Requirements

#### Dependencies
- `curl` OR `wget` (at least one required)
- `tar` (for archive extraction)
- `sha256sum` OR `shasum` (for checksum verification)

#### Supported Platforms
- **OS:** Linux, macOS
- **Architecture:** amd64 (x86_64), arm64 (aarch64)

#### Installation Prerequisites
- Write permissions to `INSTALL_DIR` (or use `sudo`)
- Network access to GitHub (github.com, api.github.com)

### Installation Process

1. Detects OS and CPU architecture
2. Validates dependencies (curl/wget, tar, checksum tools)
3. Fetches latest version from GitHub API (if not specified)
4. Downloads release archive and checksums.txt
5. Verifies SHA256 checksum of archive
6. Extracts binary from archive
7. Moves binary to `INSTALL_DIR`
8. Sets executable permissions
9. Verifies installation and shows version

### Troubleshooting

#### "Either curl or wget is required"

```bash
# macOS - Install curl via Homebrew
brew install curl

# Ubuntu/Debian - Install wget
sudo apt-get install wget

# Rocky/CentOS
sudo dnf install curl
```

#### "Cannot write to /usr/local/bin"

Use one of these solutions:

```bash
# Option 1: Use sudo
sudo ./scripts/install.sh

# Option 2: Install to user directory (recommended)
INSTALL_DIR=~/.local/bin ./scripts/install.sh

# Option 3: Create writable directory
mkdir -p ~/bin
./scripts/install.sh --dir ~/bin
export PATH="$HOME/bin:$PATH"
```

#### "Archive not found in releases"

Script failed to download archive. Check:
- Network connectivity to github.com
- GitHub API rate limiting (429 errors)
- Specified version exists: https://github.com/cesc1802/migration-tool/releases

Try specifying a known version:
```bash
./scripts/install.sh --version v0.0.3
```

#### Path Issues After Installation

If `migrate-tool` command not found:

```bash
# Check if installed
ls -l /usr/local/bin/janus

# Add to PATH if needed
export PATH="/usr/local/bin:$PATH"

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
```

### Verification

After installation:

```bash
# Check version
janus version

# Show help
janus --help
```

### Direct Binary Download

If you prefer manual installation:

1. Visit [GitHub Releases](https://github.com/cesc1802/migration-tool/releases)
2. Download binary for your platform (e.g., `migrate-tool_0.0.3_linux_amd64.tar.gz`)
3. Download `checksums.txt` from same release
4. Verify checksum:
   ```bash
   sha256sum -c checksums.txt --ignore-missing
   ```
5. Extract and install:
   ```bash
   tar -xzf migrate-tool_0.0.3_linux_amd64.tar.gz
   sudo mv janus /usr/local/bin/
   sudo chmod +x /usr/local/bin/janus
   ```

### CI/CD Integration

For automated deployments:

```bash
#!/bin/bash
# Download and install latest version
curl -fsSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | bash

# Or with specific version
curl -fsSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | bash -s -- --version v0.0.3
```

### Build from Source

Alternative to using the install script:

```bash
git clone https://github.com/cesc1802/migration-tool.git
cd migration-tool
go install ./cmd/janus
```

Requires: Go 1.25.1 or higher

---

## Windows Installation Script

**Phase 02 Completion:** Automated installation script for Windows (PowerShell).

### Overview

The `scripts/install.ps1` script provides a robust method to download and install `migrate-tool` from GitHub releases on Windows. It includes:
- Automatic architecture detection (amd64, arm64)
- SHA256 checksum verification
- Tar extraction with fallback for older Windows versions
- Optional automatic PATH configuration
- Version-specific or latest release installation

### Usage

#### Basic Installation (Latest Version)

```powershell
.\install.ps1
```

Installs latest `migrate-tool` release to `%LOCALAPPDATA%\migrate-tool`.

#### Install and Add to PATH

```powershell
.\install.ps1 -AddToPath
```

Installation directory will be added to user PATH environment variable (requires terminal restart).

#### Install Specific Version

```powershell
.\install.ps1 -Version v0.0.4
```

#### Custom Installation Directory

```powershell
.\install.ps1 -InstallDir "C:\Tools\migrate-tool"
```

#### Show Help

```powershell
.\install.ps1 -Help
```

### Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `-Version` | Install specific version | `.\install.ps1 -Version v0.0.4` |
| `-InstallDir` | Custom install directory | `.\install.ps1 -InstallDir C:\Tools` |
| `-AddToPath` | Add to user PATH (requires restart) | `.\install.ps1 -AddToPath` |
| `-Help` | Show help message | `.\install.ps1 -Help` |

### Examples

```powershell
# Install latest to default location
.\install.ps1

# Install specific version with PATH
.\install.ps1 -Version v0.0.4 -AddToPath

# Install to custom directory
.\install.ps1 -InstallDir "C:\Program Files\migrate-tool"

# Show help
.\install.ps1 -Help
```

### Requirements

#### Supported Platforms
- **OS:** Windows 10 (1803+), Windows 11, Windows Server 2016+
- **Architecture:** amd64 (x86_64), arm64 (aarch64)

#### Dependencies
- PowerShell 5.0+ (built-in on Windows 10+)
- .NET 4.6.1+ (for tar extraction fallback on older Windows)

#### Prerequisites
- Write permissions to install directory (or run as administrator)
- Network access to GitHub (github.com, api.github.com)

### Installation Process

1. Detects Windows architecture (amd64/arm64)
2. Fetches latest version from GitHub API (if not specified)
3. Downloads release archive and checksums.txt
4. Verifies SHA256 checksum of archive
5. Extracts binary using tar (with .NET fallback)
6. Creates install directory if needed
7. Installs binary to directory
8. Optionally adds directory to user PATH
9. Verifies installation and shows version

### Tar Extraction

The script uses the native `tar` command available on Windows 10 1803+ for fast extraction. For older Windows versions, it falls back to .NET-based manual tar extraction with gzip decompression.

### Troubleshooting

#### "Unsupported architecture"

Script only supports amd64 and arm64 Windows. Older 32-bit (x86) Windows is not supported.

#### "Failed to download from URL"

Check:
- Network connectivity to github.com and api.github.com
- GitHub API rate limiting (if fetching latest version)
- Specified version exists: https://github.com/cesc1802/migration-tool/releases

Try specifying a known version:
```powershell
.\install.ps1 -Version v0.0.4
```

#### "Cannot write to installation directory"

Run as administrator or choose a user-writable directory:

```powershell
# Option 1: Run as administrator
# Right-click PowerShell → Run as Administrator
# Then run: .\install.ps1

# Option 2: Install to user-local directory
.\install.ps1 -InstallDir "$env:LOCALAPPDATA\Tools\migrate-tool"

# Option 3: Install to Program Files (requires admin)
$AdminCheck = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
if (-not $AdminCheck) {
    Write-Host "Please run as Administrator"
    exit 1
}
.\install.ps1 -InstallDir "C:\Program Files\migrate-tool"
```

#### Command not found after installation

If `migrate-tool` command is not found, add the directory to PATH:

```powershell
# Check if installed
dir $env:LOCALAPPDATA\migrate-tool

# Add to PATH manually
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$newPath = "$currentPath;$env:LOCALAPPDATA\migrate-tool"
[Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

# Restart terminal for changes to take effect
```

Or use `-AddToPath` parameter during installation:

```powershell
.\install.ps1 -AddToPath
```

### Security Considerations

#### Checksum Verification

All releases include `checksums.txt` with SHA256 hashes. The script automatically verifies checksums before installation.

#### HTTPS Only

- GitHub API and release downloads use HTTPS
- Secure download with error handling

#### PowerShell Execution Policy

If you see "cannot be loaded because running scripts is disabled", enable script execution:

```powershell
# Check current policy
Get-ExecutionPolicy

# Allow current user to run scripts (no admin needed)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Then run install
.\install.ps1
```

### Verification

After installation:

```powershell
# Check version
migrate-tool version

# Show help
migrate-tool --help
```

### CI/CD Integration

For automated deployments on Windows:

```powershell
# Download and install latest version
powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1' -OutFile install.ps1; .\install.ps1"

# Or with specific version
powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1' -OutFile install.ps1; .\install.ps1 -Version v0.0.4"
```

---

## Distribution Strategy

### Supported Installation Methods

| Method | OS Support | Effort | Reliability |
|--------|-----------|--------|------------|
| Download Binary | Linux, macOS, Windows | None | High |
| Install Script (Bash) | Linux, macOS | Low | High |
| Install Script (PowerShell) | Windows | Low | High |
| `go install` | All | Low | High |
| Source Build | All | Medium | Very High |

### Script Maintenance

- **Unix Script Location:** `./scripts/install.sh`
  - **Last Updated:** Phase 01 completion
  - **Tested Platforms:** Linux (x86_64, arm64), macOS (x86_64, arm64)

- **Windows Script Location:** `./scripts/install.ps1`
  - **Last Updated:** Phase 02 completion
  - **Tested Platforms:** Windows 10 1803+ (amd64, arm64)

---

## Security Considerations

### Checksum Verification

All releases include `checksums.txt` with SHA256 hashes. The install script automatically verifies checksums before installation.

### HTTPS Only

- GitHub API and release downloads use HTTPS
- No fallback to HTTP
- Curl/wget configured with `-fsSL` (fail on HTTP error, show errors, silent, follow redirects)

### Minimal Dependencies

- No external tools beyond standard Unix utilities (curl/wget, tar, sha256sum)
- No package manager dependencies
- No requirement for sudo for user-local installations

---

## Version Management

### Fetching Latest Version

The script uses GitHub API to determine the latest release:

```bash
curl -fsSL https://api.github.com/repos/cesc1802/migration-tool/releases/latest
```

### Version Format

Versions follow semantic versioning: `vX.Y.Z` or `vX.Y.Z-suffix`

### Pinning Versions

For reproducible deployments, always specify a version:

```bash
./scripts/install.sh --version v0.0.3
```

---

## Uninstallation

```bash
# System-wide installation
sudo rm /usr/local/bin/janus

# User-local installation
rm ~/.local/bin/janus
```

---

## See Also

- [README.md](../README.md) - Quick start guide
- [Configuration Guide](./configuration.md) - Database configuration
- [Architecture Documentation](./system-architecture.md) - System design
