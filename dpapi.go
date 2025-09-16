package main

// Adapted from https://github.com/billgraziano/dpapi

import (
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

var (
	dllcrypt32 = windows.NewLazySystemDLL("Crypt32.dll")

	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(d []byte) *dataBlob {
	if len(d) == 0 {
		return &dataBlob{}
	}
	return &dataBlob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *dataBlob) toByteArray() []byte {
	d := make([]byte, b.cbData)
	/* #nosec# G103 */
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func (b *dataBlob) zeroMemory() {
	zeros := make([]byte, b.cbData)
	/* #nosec# G103 */
	copy((*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:], zeros)
}

func (b *dataBlob) free() error {
	/* #nosec# G103 */
	_, err := windows.LocalFree(windows.Handle(unsafe.Pointer(b.pbData)))
	if err != nil {
		return errors.Wrap(err, "localfree")
	}

	return nil
}

// DecryptBytes decrypts a byte array returning a byte array
func DecryptBytes(data []byte) ([]byte, error) {
	var (
		outblob dataBlob
		r       uintptr
		err     error
	)
	/* #nosec# G103 */
	r, _, err = procDecryptData.Call(
		uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, uintptr(0x0),
		uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, errors.Wrap(err, "procdecryptdata")
	}

	dec := outblob.toByteArray()
	outblob.zeroMemory()
	return dec, outblob.free()
}
