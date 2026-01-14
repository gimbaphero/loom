# Chocolatey Package for Loom

Chocolatey is the most popular package manager for Windows, with over 9,000 community-maintained packages.

## Installation (for users)

Once published to Chocolatey.org:

```powershell
# Install Loom
choco install loom

# Update Loom
choco upgrade loom

# Uninstall Loom
choco uninstall loom
```

## Package Structure

```
chocolatey/
├── loom.nuspec                    # Package metadata (XML)
├── LICENSE.txt                    # Apache 2.0 license
├── tools/
│   ├── chocolateyinstall.ps1     # Installation script
│   └── chocolateyuninstall.ps1   # Uninstallation script
└── README.md                      # This file
```

## Building the Package (for maintainers)

### Prerequisites

```powershell
# Install Chocolatey (if not already installed)
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install chocolatey package builder
choco install checksum
```

### Updating for New Releases

1. **Update version** in `loom.nuspec`:
   ```xml
   <version>1.0.1</version>
   ```

2. **Update URLs and version** in `tools/chocolateyinstall.ps1`:
   ```powershell
   $version = '1.0.1'
   ```

3. **Calculate checksums** for the release binaries:

   ```powershell
   # Download binaries
   $version = "1.0.1"
   Invoke-WebRequest "https://github.com/teradata-labs/loom/releases/download/v$version/loom-windows-amd64.exe.zip" -OutFile loom.zip
   Invoke-WebRequest "https://github.com/teradata-labs/loom/releases/download/v$version/looms-windows-amd64.exe.zip" -OutFile looms.zip

   # Calculate SHA256 checksums
   checksum -t sha256 -f loom.zip
   checksum -t sha256 -f looms.zip
   ```

4. **Update checksums** in `tools/chocolateyinstall.ps1`:
   ```powershell
   checksum64    = 'ABC123...'  # First hash (loom)
   # and
   checksum64    = 'DEF456...'  # Second hash (looms)
   ```

### Testing Locally

```powershell
# Navigate to chocolatey directory
cd chocolatey

# Pack the package
choco pack

# Test installation locally
choco install loom -source . -y

# Test the binaries
loom --help
looms --help

# Uninstall
choco uninstall loom -y
```

### Publishing to Chocolatey.org

1. **Create account** at https://community.chocolatey.org/account/Register

2. **Get API key** from https://community.chocolatey.org/account

3. **Set API key** (one-time setup):
   ```powershell
   choco apikey --key YOUR_API_KEY --source https://push.chocolatey.org/
   ```

4. **Build and push**:
   ```powershell
   # Build package
   choco pack

   # Push to Chocolatey
   choco push loom.1.0.1.nupkg --source https://push.chocolatey.org/
   ```

5. **Wait for moderation**: Community packages require manual approval (usually 1-3 days)

## Package Features

### What Gets Installed

- **Binaries**: `loom.exe` and `looms.exe` added to PATH via shims
- **Patterns**: 90+ YAML patterns downloaded to `$env:USERPROFILE\.loom\patterns\`
- **Environment Variable**: `LOOM_DATA_DIR` set to `$env:USERPROFILE\.loom`
- **Configuration**: Empty `looms.yaml` created (user configures LLM provider post-install)

### What Gets Uninstalled

- **Binaries**: Removed from PATH
- **Environment Variable**: `LOOM_DATA_DIR` removed

**Preserved** (user data):
- `$env:USERPROFILE\.loom\` directory (patterns, config, database)
- Users can manually remove with: `Remove-Item -Path "$env:USERPROFILE\.loom" -Recurse -Force`

## Chocolatey Guidelines

- **Title Case**: Package title uses proper casing ("Loom AI Agent Framework")
- **Tags**: Space-separated, lowercase, relevant keywords
- **Description**: Markdown supported, comprehensive but concise
- **Dependencies**: None (standalone package)
- **License**: Must match upstream license (Apache 2.0)
- **Binaries**: Downloaded from official GitHub releases only

## Testing Checklist

Before publishing:

- [ ] Package builds without errors: `choco pack`
- [ ] Installation works: `choco install loom -source . -y`
- [ ] Binaries are in PATH: `where loom` and `where looms`
- [ ] Binaries execute: `loom --help` and `looms --help`
- [ ] Patterns installed: Check `$env:USERPROFILE\.loom\patterns\`
- [ ] Environment variable set: `$env:LOOM_DATA_DIR`
- [ ] Uninstall works: `choco uninstall loom -y`
- [ ] Upgrade works: `choco upgrade loom -source . -y`

## Troubleshooting

### "Package already exists"
```powershell
# Remove old package
Remove-Item *.nupkg
choco pack
```

### "Checksum mismatch"
Recalculate checksums:
```powershell
checksum -t sha256 -f path/to/binary.zip
```

### "Cannot find package"
Make sure you're in the chocolatey directory:
```powershell
Get-Location  # Should show .../loom/chocolatey
```

## References

- [Chocolatey Documentation](https://docs.chocolatey.org/)
- [Package Guidelines](https://docs.chocolatey.org/en-us/create/create-packages)
- [Helper Functions](https://docs.chocolatey.org/en-us/create/functions/)
- [Moderation Process](https://docs.chocolatey.org/en-us/community-repository/moderation)
