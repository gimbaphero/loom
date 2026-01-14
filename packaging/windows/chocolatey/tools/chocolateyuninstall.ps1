$ErrorActionPreference = 'Stop'

$packageName = 'loom'

# Remove shims
Uninstall-BinFile -Name 'loom'
Uninstall-BinFile -Name 'looms'

# Remove environment variable
Uninstall-ChocolateyEnvironmentVariable -VariableName 'LOOM_DATA_DIR' -VariableType 'User'

Write-Host "Loom has been uninstalled." -ForegroundColor Green
Write-Host ""
Write-Host "Note: The following were NOT removed (manual cleanup required):" -ForegroundColor Yellow
Write-Host "  • Loom data directory: $env:USERPROFILE\.loom"
Write-Host "  • Patterns: $env:USERPROFILE\.loom\patterns"
Write-Host "  • Configuration: $env:USERPROFILE\.loom\looms.yaml"
Write-Host "  • Database: $env:USERPROFILE\.loom\loom.db"
Write-Host ""
Write-Host "To remove all Loom data:" -ForegroundColor Cyan
Write-Host "  Remove-Item -Path `"$env:USERPROFILE\.loom`" -Recurse -Force"
Write-Host ""
