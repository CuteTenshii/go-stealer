package main

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/sys/windows/registry"
)

// GetRAMUsage returns the current RAM usage in MB.
func GetRAMUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Sys / 1024 / 1024 // Convert bytes to MB
}

func GetCpuName() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		panic(err)
	}
	defer key.Close()

	cpuName, _, err := key.GetStringValue("ProcessorNameString")
	if err != nil {
		panic(err)
	}

	return cpuName
}

// GetFullOSName returns the full name of the operating system. For example, "Windows 10 Pro (Build 19042)".
func GetFullOSName() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		panic(err)
	}
	defer key.Close()

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		panic(err)
	}

	buildNumber, _, err := key.GetStringValue("CurrentBuildNumber")
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s (Build %s)", productName, buildNumber)
}

func GetComputerName() string {
	return os.Getenv("COMPUTERNAME")
}
