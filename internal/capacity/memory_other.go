//go:build !linux && !windows

package capacity

func detectMemoryMB() (total, avail int) {
	return 0, 0
}
