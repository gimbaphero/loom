# Scoop Manifests for Loom

Scoop is a command-line installer for Windows. These manifests allow users to install Loom easily with `scoop install`.

## Installation (for users)

### Option 1: Install from this bucket (future)

Once submitted to a Scoop bucket:

```powershell
# Install Loom TUI client
scoop install loom

# Install Loom server
scoop install loom-server

# Or install both
scoop install loom loom-server
```

### Option 2: Install directly from this repository

```powershell
# Install from local manifest
scoop install https://raw.githubusercontent.com/teradata-labs/loom/main/scoop/loom.json
scoop install https://raw.githubusercontent.com/teradata-labs/loom/main/scoop/loom-server.json
```

## Maintenance (for maintainers)

### Updating version and hash

When a new version is released:

1. Update the `version` field in both manifests
2. Calculate SHA256 hashes for the release binaries:

```powershell
# Download the release binaries
$version = "1.0.0"
Invoke-WebRequest "https://github.com/teradata-labs/loom/releases/download/v$version/loom-windows-amd64.exe.zip" -OutFile loom.zip
Invoke-WebRequest "https://github.com/teradata-labs/loom/releases/download/v$version/looms-windows-amd64.exe.zip" -OutFile looms.zip

# Calculate hashes
Get-FileHash loom.zip -Algorithm SHA256 | Select-Object -ExpandProperty Hash
Get-FileHash looms.zip -Algorithm SHA256 | Select-Object -ExpandProperty Hash
```

3. Update the `hash` fields in the manifests
4. Test installation:

```powershell
scoop install .\loom.json
scoop install .\loom-server.json
```

### Submitting to Scoop Buckets

#### Option 1: Submit to scoop-extras (recommended)

The `scoop-extras` bucket is for GUI apps and complex installations:

1. Fork https://github.com/ScoopInstaller/Extras
2. Copy `loom.json` to `bucket/loom.json`
3. Copy `loom-server.json` to `bucket/loom-server.json`
4. Submit a pull request

#### Option 2: Create your own bucket

```powershell
# Create a bucket repository on GitHub (e.g., teradata-labs/scoop-loom)
# Then users can add it:
scoop bucket add loom https://github.com/teradata-labs/scoop-loom
scoop install loom/loom
scoop install loom/loom-server
```

## Manifest Structure

- **version**: Current version (must match GitHub release tag)
- **url**: Download URL for the release binary
- **hash**: SHA256 hash of the download file
- **bin**: Name of the executable to add to PATH
- **checkver**: How to check for new versions
- **autoupdate**: How to automatically update the manifest
- **post_install**: PowerShell commands to run after installation (server only)
- **notes**: Instructions shown to users after installation

## Testing

```powershell
# Test local manifest
scoop install .\loom.json

# Test binary works
loom-windows-amd64 --help

# Uninstall
scoop uninstall loom
```

## References

- [Scoop Documentation](https://github.com/ScoopInstaller/Scoop/wiki)
- [App Manifests](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifests)
- [Autoupdate](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifest-Autoupdate)
