// win_screenshot.go
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procGetDC            = user32.NewProc("GetDC")
	procReleaseDC        = user32.NewProc("ReleaseDC")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
)

const (
	SM_XVIRTUALSCREEN  = 76
	SM_YVIRTUALSCREEN  = 77
	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79

	SRCCOPY = 0x00CC0020
)

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

func getSystemMetric(index int) int {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(ret)
}

func TakeScreenshot() []byte {
	// Virtual screen bounds (covers all monitors)
	x := getSystemMetric(SM_XVIRTUALSCREEN)
	y := getSystemMetric(SM_YVIRTUALSCREEN)
	w := getSystemMetric(SM_CXVIRTUALSCREEN)
	h := getSystemMetric(SM_CYVIRTUALSCREEN)
	if w == 0 || h == 0 {
		fmt.Println("no screen found")
		return nil
	}
	fmt.Printf("Virtual screen: x=%d y=%d w=%d h=%d\n", x, y, w, h)

	// hdcScreen = GetDC(NULL)
	hdcScreen, _, _ := procGetDC.Call(0)
	if hdcScreen == 0 {
		fmt.Println("GetDC failed")
		return nil
	}
	defer procReleaseDC.Call(0, hdcScreen)

	// hdcMem = CreateCompatibleDC(hdcScreen)
	hdcMem, _, _ := procCreateCompatibleDC.Call(hdcScreen)
	if hdcMem == 0 {
		fmt.Println("CreateCompatibleDC failed")
		return nil
	}
	defer procDeleteDC.Call(hdcMem)

	// hBitmap = CreateCompatibleBitmap(hdcScreen, w, h)
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdcScreen, uintptr(w), uintptr(h))
	if hBitmap == 0 {
		fmt.Println("CreateCompatibleBitmap failed")
		return nil
	}
	defer procDeleteObject.Call(hBitmap)

	// Select bitmap into memory DC
	prevObj, _, _ := procSelectObject.Call(hdcMem, hBitmap)
	if prevObj == 0 {
		fmt.Println("SelectObject failed")
		return nil
	}
	// BitBlt from screen -> mem
	ret, _, _ := procBitBlt.Call(
		hdcMem,
		0, 0,
		uintptr(w), uintptr(h),
		hdcScreen,
		uintptr(x), uintptr(y),
		uintptr(SRCCOPY))
	if ret == 0 {
		fmt.Println("BitBlt failed")
		return nil
	}

	// Prepare BITMAPINFO for GetDIBits
	var bih BITMAPINFOHEADER
	bih.BiSize = uint32(unsafe.Sizeof(bih))
	bih.BiWidth = int32(w)
	bih.BiHeight = int32(h) // positive -> bottom-up DIB
	bih.BiPlanes = 1
	bih.BiBitCount = 32   // request 32-bit ARGB/BGRA
	bih.BiCompression = 0 // BI_RGB
	// image buffer (4 bytes per pixel)
	rowBytes := w * 4
	imgSize := rowBytes * h
	buf := make([]byte, imgSize)

	// Build BITMAPINFO: BITMAPINFOHEADER + color table (none for 32bpp)
	// We will pass pointer to BITMAPINFOHEADER directly.
	ret2, _, _ := procGetDIBits.Call(
		hdcScreen,
		hBitmap,
		0,
		uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bih)),
		0) // DIB_RGB_COLORS = 0
	if ret2 == 0 {
		return nil
	}

	// Convert BGRA byte buffer to Go image.RGBA
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// The DIB returned by GetDIBits for bottom-up bitmaps is bottom-up:
	// iterate rows from bottom to top.
	for row := 0; row < h; row++ {
		srcRow := (h - 1 - row) * rowBytes
		dstRow := row * img.Stride
		copy(img.Pix[dstRow:dstRow+rowBytes], buf[srcRow:srcRow+rowBytes])
		// Windows returns BGRA: reorder to RGBA
		for col := 0; col < w; col++ {
			i := dstRow + col*4
			b := img.Pix[i+0]
			g := img.Pix[i+1]
			r := img.Pix[i+2]
			a := img.Pix[i+3]
			img.Pix[i+0] = r
			img.Pix[i+1] = g
			img.Pix[i+2] = b
			img.Pix[i+3] = a
		}
	}

	var f bytes.Buffer
	if err := png.Encode(&f, img); err != nil {
		return nil
	}

	return f.Bytes()
}
