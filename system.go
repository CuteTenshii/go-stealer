package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

type MEMORYSTATUSEX struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// GetRAMTotal returns the total physical RAM in GB
func GetRAMTotal() float64 {
	var memStatus MEMORYSTATUSEX
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))

	// Call GlobalMemoryStatusEx
	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		panic(err)
	}

	totalRAM := memStatus.TotalPhys // in bytes
	return float64(totalRAM) / (1024 * 1024 * 1024)
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
