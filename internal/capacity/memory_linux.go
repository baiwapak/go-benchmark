//go:build linux

package capacity

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
)

func detectMemoryMB() (total, avail int) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			total = parseKBLine(line)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			avail = parseKBLine(line)
		}
	}
	if avail == 0 && total > 0 {
		avail = total / 2
	}
	return total, avail
}

func parseKBLine(line string) int {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	kb, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0
	}
	return kb / 1024
}
