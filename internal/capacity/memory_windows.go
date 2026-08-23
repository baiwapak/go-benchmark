//go:build windows

package capacity

import (
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

func detectMemoryMB() (total, avail int) {
	var ms memoryStatusEx
	ms.length = uint32(unsafe.Sizeof(ms))
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GlobalMemoryStatusEx")
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&ms)))
	if r == 0 {
		return 0, 0
	}
	return int(ms.totalPhys / 1024 / 1024), int(ms.availPhys / 1024 / 1024)
}
