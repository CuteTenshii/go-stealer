package main

import (
	"syscall"
	"unsafe"
)

var (
	procMessageBoxW = user32.NewProc("MessageBoxW")
)

const (
	MB_ICONERROR = 0x00000010
)

func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	captionPtr, _ := syscall.UTF16PtrFromString(caption)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	ret, _, _ := procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(captionPtr)), uintptr(unsafe.Pointer(titlePtr)), uintptr(flags))
	return int(ret)
}
