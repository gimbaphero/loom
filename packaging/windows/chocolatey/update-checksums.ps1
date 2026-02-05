# Update chocolateyinstall.ps1 with real checksums from GitHub releases
# Usage: .\update-checksums.ps1 -Version "1.0.2"

param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

$ErrorActionPreference = 'Stop'

# GitHub release URLs
$baseUrl = "https://github.com/teradata-labs/loom/releases/download/v$Version"
$loomChecksumUrl = "$baseUrl/loom-windows-amd64.exe.zip.sha256"
$loomsChecksumUrl = "$baseUrl/looms-windows-amd64.exe.zip.sha256"

Write-Host "Fetching checksums for version $Version..." -ForegroundColor Cyan

try {
    # Fetch checksums (decode byte array as UTF-8)
    $loomResponse = Invoke-WebRequest -Uri $loomChecksumUrl -UseBasicParsing
    $loomChecksumRaw = [System.Text.Encoding]::UTF8.GetString($loomResponse.Content).Trim()
    # Handle both "hash" and "hash  filename" formats - split and take first part
    $loomChecksum = ($loomChecksumRaw -split '\s+')[0]

    # Validate checksum is exactly 64 hex characters
    if ($loomChecksum.Length -ne 64 -or $loomChecksum -notmatch '^[0-9a-fA-F]{64}$') {
        throw "Invalid loom checksum: '$loomChecksum' (length: $($loomChecksum.Length))"
    }

    $loomsResponse = Invoke-WebRequest -Uri $loomsChecksumUrl -UseBasicParsing
    $loomsChecksumRaw = [System.Text.Encoding]::UTF8.GetString($loomsResponse.Content).Trim()
    # Handle both "hash" and "hash  filename" formats - split and take first part
    $loomsChecksum = ($loomsChecksumRaw -split '\s+')[0]

    # Validate checksum is exactly 64 hex characters
    if ($loomsChecksum.Length -ne 64 -or $loomsChecksum -notmatch '^[0-9a-fA-F]{64}$') {
        throw "Invalid looms checksum: '$loomsChecksum' (length: $($loomsChecksum.Length))"
    }

    Write-Host "✓ loom checksum: $loomChecksum" -ForegroundColor Green
    Write-Host "✓ looms checksum: $loomsChecksum" -ForegroundColor Green

    # Read install script
    $scriptPath = Join-Path $PSScriptRoot "tools\chocolateyinstall.ps1"
    $content = Get-Content $scriptPath -Raw

    # Update version (use string concatenation to avoid escaping issues)
    $content = $content -replace "(\$version = ').*?(')", ('$1' + $Version + '$2')

    # Update loom checksum (first occurrence)
    $content = $content -replace "(checksum64\s*=\s*')[0-9a-fA-F]{64}('.*?# loom TUI)", ('$1' + $loomChecksum + '$2')

    # Update looms checksum (second occurrence)
    $content = $content -replace "(\`$packageArgs\['checksum64'\] = ')[0-9a-fA-F]{64}('.*?# looms server)", ('$1' + $loomsChecksum + '$2')

    # Write updated content (keep newlines intact!)
    Set-Content -Path $scriptPath -Value $content

    Write-Host ""
    Write-Host "✅ Successfully updated chocolateyinstall.ps1" -ForegroundColor Green
    Write-Host "   Version: $Version" -ForegroundColor Yellow
    Write-Host "   Loom checksum: $loomChecksum" -ForegroundColor Yellow
    Write-Host "   Looms checksum: $loomsChecksum" -ForegroundColor Yellow

} catch {
    Write-Error "Failed to update checksums: $_"
    Write-Host ""
    Write-Host "Possible reasons:" -ForegroundColor Yellow
    Write-Host "  1. Release v$Version doesn't exist on GitHub"
    Write-Host "  2. Checksum files not uploaded yet"
    Write-Host "  3. Network connectivity issues"
    exit 1
}
