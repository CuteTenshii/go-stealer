package anti_vm

import (
	"fmt"
	"strings"
	"unsafe"
)

var (
	getSystemFirmwareTable = kernel32.NewProc("GetSystemFirmwareTable")
)

// SMBIOS table types
const (
	SMBIOS_TYPE_BIOS      = 0
	SMBIOS_TYPE_SYSTEM    = 1
	SMBIOS_TYPE_PROCESSOR = 4
)

// SMBIOS string offsets (in formatted sections)
const (
	BIOS_VENDOR_INDEX_OFFSET   = 4    // Type 0
	SYSTEM_MANUFACTURER_OFFSET = 4    // Type 1
	PROCESSOR_VERSION_OFFSET   = 0x10 // Type 4
)

// SMBIOS signature for GetSystemFirmwareTable
const (
	RSMB_SIGNATURE = 0x52534D42 // 'RSMB'
)

type SMBIOSInfo struct {
	// Manufacturer represents the system manufacturer. Example: "Dell Inc."
	Manufacturer string
	// BiosManufacturer represents the BIOS manufacturer. Example: "Dell Inc."
	BiosManufacturer string
	// CPUModel represents the CPU model. Example: "Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz"
	CPUModel string
}

func getSMBIOSTable() ([]byte, error) {
	// First call to get the required buffer size
	ret, _, err := getSystemFirmwareTable.Call(uintptr(RSMB_SIGNATURE), uintptr(0), uintptr(0), uintptr(0))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get SMBIOS size: %v", err)
	}

	buf := make([]byte, ret)
	ret, _, err = getSystemFirmwareTable.Call(uintptr(RSMB_SIGNATURE), uintptr(0), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get SMBIOS table: %v", err)
	}

	return buf, nil
}

// parseString returns the N-th string after the formatted section
func parseString(strings []string, index int) string {
	if index <= 0 || index > len(strings) {
		return ""
	}
	return strings[index-1]
}

func parseSMBIOS(buf []byte) SMBIOSInfo {
	var info SMBIOSInfo
	i := 0
	for i+4 < len(buf) {
		t := buf[i]   // Type
		l := buf[i+1] // Length
		if int(l) > len(buf)-i {
			break // malformed entry, prevent panic
		}
		data := buf[i : i+int(l)]

		j := i + int(l)
		strStart := j
		var strList []string
		for strStart < len(buf) {
			if buf[strStart] == 0 {
				if strStart+1 < len(buf) && buf[strStart+1] == 0 {
					j = strStart + 2
					break
				}
			}
			strStart++
		}
		rawStrings := buf[i+int(l) : j-2]
		for _, s := range strings.Split(string(rawStrings), "\x00") {
			if s != "" {
				strList = append(strList, s)
			}
		}

		switch t {
		case SMBIOS_TYPE_BIOS:
			if len(data) > BIOS_VENDOR_INDEX_OFFSET {
				idx := int(data[BIOS_VENDOR_INDEX_OFFSET])
				info.BiosManufacturer = strList[idx-1]
			}
		case SMBIOS_TYPE_SYSTEM:
			if len(data) > SYSTEM_MANUFACTURER_OFFSET {
				idx := int(data[SYSTEM_MANUFACTURER_OFFSET])
				info.Manufacturer = parseString(strList, idx)
			}
		case SMBIOS_TYPE_PROCESSOR:
			if len(data) > PROCESSOR_VERSION_OFFSET {
				idx := int(data[PROCESSOR_VERSION_OFFSET])
				info.CPUModel = parseString(strList, idx)
			}
		}

		i = j
	}

	return info
}

func CheckSMBIOS() bool {
	buf, err := getSMBIOSTable()
	if err != nil {
		return false
	}

	info := parseSMBIOS(buf)
	info.Manufacturer = strings.ToLower(info.Manufacturer)
	info.BiosManufacturer = strings.ToLower(info.BiosManufacturer)
	info.CPUModel = strings.ToLower(info.CPUModel)

	suspiciousManufacturers := []string{
		"vmware", "xen", "qemu", "bochs", "parallels", "virtualbox", "kvm", "microsoft corporation",
		"amazon", "google", "digitalocean", "hetzner", "oracle", "alibaba",
	}
	// Those CPU models are often found in VMs or cloud instances.
	suspiciousCPUModels := []string{
		"intel(r) xeon(r)", "amd epyc", "intel(r) atom(tm)", "intel(r) core(tm) i3-8100t", "intel(r) core(tm) i5-8259u",
	}

	for _, sm := range suspiciousManufacturers {
		sm = strings.ToLower(sm)
		// Check Manufacturer and BIOS Manufacturer
		if strings.Contains(info.Manufacturer, sm) || strings.Contains(info.BiosManufacturer, sm) {
			return true
		}
	}

	for _, cpu := range suspiciousCPUModels {
		cpu = strings.ToLower(cpu)
		if strings.Contains(info.CPUModel, cpu) {
			return true
		}
	}

	return false
}
