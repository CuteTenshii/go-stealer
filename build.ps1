function Check-BinaryExists
{
    param ([string]$binary)
    $null -ne (Get-Command $binary -ErrorAction SilentlyContinue)
}

if (-not (Check-BinaryExists -binary "go"))
{
    Write-Host "Go is required to build this software, but it is not installed."

    # Check if winget is available
    if (Check-BinaryExists -binary "winget")
    {
        $choice = Read-Host "WinGet is available. Do you want to automatically install Go? (Y/N)"
        if ($choice -match "^[Yy]$")
        {
            Write-Host "Installing Go via WinGet..."
            Start-Process -FilePath winget -ArgumentList 'install -e --id GoLang.Go --silent --accept-package-agreements --accept-source-agreements' -NoNewWindow -Wait
            Write-Host "Go installation finished. Please restart the script and run this script again."
            exit 0
        } else {
            Write-Host "Please install Go manually from https://go.dev/dl/"
            exit 1
        }
    } else {
        Write-Host "Winget not found. Please install Go manually from https://go.dev/dl/"
        exit 1
    }
}

# Check if a C++ compiler is installed (for CGO).
if (-not (Check-BinaryExists -binary "gcc"))
{
    Write-Host "GCC (MinGW) is required to build this software with CGO support, but it is not installed."
    $choice = Read-Host "This script can automatically install it if you want. Proceed? (Y/N)"
    if ($choice -match "^[Yy]$")
    {
        Write-Host "Installing MSYS2 and GCC..."
        Write-Host "This will download and install MSYS2, which may take some time."

        # Download MSYS2 installer
        Invoke-WebRequest -Uri "https://github.com/msys2/msys2-installer/releases/download/nightly-x86_64/msys2-x86_64-latest.exe" -OutFile "msys2.exe"

        # Install silently (adjust path as needed)
        Write-Host "Starting MSYS2 installation..."
        Start-Process msys2.exe -ArgumentList "install -c --al --am -t C:\msys64" -NoNewWindow -Wait -PassThru
        Remove-Item msys2.exe
        Write-Host "MSYS2 installation completed."

        Write-Host "Updating MSYS2 and installing GCC... Please wait, this should be fast."
        Start-Process -FilePath "C:\msys64\usr\bin\bash.exe" -ArgumentList "-lc 'pacman -Syu --noconfirm mingw-w64-x86_64-gcc'" -NoNewWindow -Wait -PassThru

        # Update PATH to include MSYS2 binaries
        if (-not ($env:PATH -like "*C:\msys64\mingw64\bin*"))
        {
            [System.Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\msys64\mingw64\bin", [System.EnvironmentVariableTarget]::User)
        }

        Write-Host "GCC installation finished. Please restart the script and run this script again."
        exit 0
    }
}

# Navigate to the script's directory
Set-Location -Path $PSScriptRoot

# Ensure Go modules are up to date
Write-Host "Updating Go modules..."
Start-Process -FilePath go -ArgumentList "mod", "tidy" -NoNewWindow -Wait -PassThru

# Prompt for filename
$filename = Read-Host "Enter the output filename (without extension)"
if ([string]::IsNullOrWhiteSpace($filename))
{
    Write-Host "Filename cannot be empty."
    exit 1
}

$output = "$filename.exe"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

# Build the project
Write-Host "Building your executable... This may be fast or take a while depending on your system."
$build = Start-Process -FilePath go -ArgumentList "build", '-ldflags="-s -w"', "-o", "$output", "." -NoNewWindow -Wait -PassThru

if ($build.ExitCode -ne 0)
{
    Write-Host "Build failed."
    exit 1
}

Write-Host "Build succeeded. Output file: $output"
