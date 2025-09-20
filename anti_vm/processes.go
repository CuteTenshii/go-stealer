package anti_vm

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	psapi    = windows.NewLazySystemDLL("psapi.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procOpenProcess       = kernel32.NewProc("OpenProcess")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procEnumProcesses     = psapi.NewProc("EnumProcesses")
	procQueryFullProcName = psapi.NewProc("GetModuleBaseNameW") // For process name
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// CheckProcesses checks for the presence of common VM-related processes, debuggers, and analysis tools.
func CheckProcesses() bool {
	const maxProcesses = 1024
	processIDs := make([]uint32, maxProcesses)
	var bytesReturned uint32

	r1, _, _ := procEnumProcesses.Call(uintptr(unsafe.Pointer(&processIDs[0])), uintptr(maxProcesses*4), uintptr(unsafe.Pointer(&bytesReturned)))
	if r1 == 0 {
		return false
	}

	// List of common VM-related processes, debuggers, and analysis tools
	suspicious := []string{
		"vboxservice", "vboxtray", "vmtoolsd", "vmwaretray", "vmwareuser", "qemu-ga", "xenservice", "xenclient",
		"ollydbg", "ida", "ida64", "idag", "idag64", "idaw", "idaw64", "idaq", "idaq64", "scylla", "scylla_x64", "scylla_x86",
		"procexp", "procexp64", "processhacker", "wireshark", "fiddler", "httpdebuggerui", "httpdebuggersvc", "qemu",
		"vmmouse", "vmusrvc", "vmsrvc",
	}

	for i := 0; i < int(bytesReturned/4); i++ {
		pid := processIDs[i]
		handle, _, _ := procOpenProcess.Call(
			uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
			0,
			uintptr(pid),
		)
		if handle == 0 {
			continue
		}

		nameBuf := make([]uint16, windows.MAX_PATH)
		ret, _, _ := procQueryFullProcName.Call(
			handle,
			0,
			uintptr(unsafe.Pointer(&nameBuf[0])),
			uintptr(len(nameBuf)),
		)
		procCloseHandle.Call(handle)

		if ret == 0 {
			continue
		}

		procName := strings.ToLower(windows.UTF16ToString(nameBuf))
		// Strip path and extension
		if idx := strings.LastIndex(procName, `\`); idx != -1 {
			procName = procName[idx+1:]
		}
		if idx := strings.LastIndex(procName, "."); idx != -1 {
			procName = procName[:idx]
		}

		for _, s := range suspicious {
			if procName == s {
				return true
			}
		}
	}
	return false
}
