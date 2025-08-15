# Build script for BW AKA Finder
# This script creates a single executable with all dependencies embedded

# Clean previous builds
Write-Host "Cleaning previous builds..."
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
}

# Create dist directory
New-Item -ItemType Directory -Path "dist" -Force | Out-Null

# Build the application with all dependencies included
Write-Host "Building BW AKA Finder..."
$env:GOOS="windows"
$env:GOARCH="amd64"

# Build with embedded resources
go build -tags=windows -o dist/bwakafinder.exe .

# Check if build was successful
if (Test-Path "dist/bwakafinder.exe") {
    Write-Host "Build successful!" -ForegroundColor Green
    $size = (Get-Item "dist/bwakafinder.exe").Length / 1MB
    Write-Host "Executable size: $($size.ToString("F2")) MB" -ForegroundColor Cyan
    
    # Copy the executable to the root directory as well
    Copy-Item "dist/bwakafinder.exe" -Destination "bwakafinder.exe"
    Write-Host "Executable copied to root directory" -ForegroundColor Cyan
    
    Write-Host "\nBuild complete! The executable can be found in:" -ForegroundColor Green
    Write-Host "  1. dist/bwakafinder.exe" -ForegroundColor Yellow
    Write-Host "  2. bwakafinder.exe (root directory)" -ForegroundColor Yellow
    Write-Host "\nThis executable contains all dependencies and can be distributed as a standalone application." -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
