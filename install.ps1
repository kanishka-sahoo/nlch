# nlch Installation Script for Windows
# This script downloads and installs the latest release of nlch

param(
    [switch]$Help,
    [switch]$Version,
    [string]$InstallDir = "$env:USERPROFILE\bin"
)

# Configuration
$REPO = "kanishka-sahoo/nlch"
$BINARY_NAME = "nlch"

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# Helper functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
}

# Show help
function Show-Help {
    Write-Host "nlch Installation Script for Windows" -ForegroundColor $Colors.Blue
    Write-Host ""
    Write-Host "Usage: .\install.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Help          Show this help message"
    Write-Host "  -Version       Show script version"
    Write-Host "  -InstallDir    Installation directory (default: $env:USERPROFILE\bin)"
    Write-Host ""
    Write-Host "This script downloads and installs the latest release of nlch"
    Write-Host "from the GitHub repository: https://github.com/$REPO"
}

# Show version
function Show-Version {
    Write-Host "nlch installation script v1.0.0"
}

# Detect platform
function Get-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    return "windows-$arch"
}

# Check if running in Windows Subsystem for Linux
function Test-WSL {
    return $env:WSL_DISTRO_NAME -ne $null
}

# Get latest release info from GitHub API
function Get-LatestRelease {
    $apiUrl = "https://api.github.com/repos/$REPO/releases/latest"
    
    Write-Info "Fetching latest release information..."
    
    try {
        $releaseInfo = Invoke-RestMethod -Uri $apiUrl -Method Get
        return $releaseInfo
    }
    catch {
        Write-Error "Failed to fetch release information: $($_.Exception.Message)"
        exit 1
    }
}

# Get download URL for the platform
function Get-DownloadUrl {
    param(
        [object]$ReleaseInfo,
        [string]$Platform
    )
    
    $assetName = "$BINARY_NAME-$Platform.exe"
    
    $asset = $ReleaseInfo.assets | Where-Object { $_.name -eq $assetName }
    
    if (-not $asset) {
        Write-Error "No release asset found for platform: $Platform"
        Write-Info "Available assets:"
        $ReleaseInfo.assets | Where-Object { $_.name -like "*$BINARY_NAME*" } | ForEach-Object { Write-Host "  $($_.name)" }
        exit 1
    }
    
    return $asset.browser_download_url
}

# Download and install binary
function Install-Binary {
    param(
        [string]$DownloadUrl,
        [string]$Platform
    )
    
    $tempFile = [System.IO.Path]::GetTempFileName()
    $tempFile = [System.IO.Path]::ChangeExtension($tempFile, ".exe")
    
    Write-Info "Downloading $BINARY_NAME from: $DownloadUrl"
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $tempFile
    }
    catch {
        Write-Error "Failed to download binary: $($_.Exception.Message)"
        exit 1
    }
    
    if (-not (Test-Path $tempFile)) {
        Write-Error "Failed to download binary"
        exit 1
    }
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # Install binary
    $installPath = Join-Path $InstallDir "$BINARY_NAME.exe"
    Write-Info "Installing to: $installPath"
    
    try {
        Copy-Item $tempFile $installPath -Force
        Remove-Item $tempFile -Force
        Write-Success "$BINARY_NAME installed successfully to $installPath"
    }
    catch {
        Write-Error "Failed to install binary: $($_.Exception.Message)"
        exit 1
    }
}

# Verify installation
function Test-Installation {
    $binaryPath = Join-Path $InstallDir "$BINARY_NAME.exe"
    
    if (Test-Path $binaryPath) {
        try {
            $version = & $binaryPath --version 2>$null
            if (-not $version) { $version = "unknown" }
            Write-Success "Installation verified. Version: $version"
        }
        catch {
            Write-Success "Binary installed successfully"
        }
        
        # Check if install directory is in PATH
        $pathDirs = $env:PATH -split ';'
        if ($pathDirs -contains $InstallDir) {
            Write-Info "You can now use '$BINARY_NAME' from anywhere in your terminal"
        }
        else {
            Write-Warning "Install directory is not in your PATH"
            Write-Info "To add to PATH, run this command in an elevated PowerShell:"
            Write-Info "  [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$InstallDir', 'User')"
            Write-Info "Or restart your terminal and run: $binaryPath"
        }
    }
    else {
        Write-Error "Installation verification failed"
        exit 1
    }
}

# Main installation process
function Start-Installation {
    Write-Info "Starting nlch installation..."
    
    # Detect platform
    $platform = Get-Platform
    Write-Info "Detected platform: $platform"
    
    # Get latest release
    $releaseInfo = Get-LatestRelease
    Write-Info "Latest version: $($releaseInfo.tag_name)"
    
    # Get download URL
    $downloadUrl = Get-DownloadUrl -ReleaseInfo $releaseInfo -Platform $platform
    
    # Install binary
    Install-Binary -DownloadUrl $downloadUrl -Platform $platform
    
    # Verify installation
    Test-Installation
    
    Write-Success "Installation complete!"
    Write-Info "Run '$BINARY_NAME --help' to get started"
}

# Handle script parameters
if ($Help) {
    Show-Help
    exit 0
}

if ($Version) {
    Show-Version
    exit 0
}

# Check PowerShell execution policy
$executionPolicy = Get-ExecutionPolicy
if ($executionPolicy -eq "Restricted") {
    Write-Warning "PowerShell execution policy is set to 'Restricted'"
    Write-Info "You may need to run: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser"
    Write-Info "Or run this script with: powershell -ExecutionPolicy Bypass -File install.ps1"
}

# Run main installation
try {
    Start-Installation
}
catch {
    Write-Error "Installation failed: $($_.Exception.Message)"
    exit 1
}
