package sampler

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Mem struct {
	TotalKB         uint64
	AvailableKB     uint64
	SwapTotalKB     uint64
	SwapAvailableKB uint64
}

// Memory used percentage
func (m Mem) UsedPercent() float64 {
	if m.TotalKB == 0 {
		return 0
	}

	return float64(m.TotalKB-m.AvailableKB) / float64(m.TotalKB) * 100

}

// Swap Memory used percentage
func (m Mem) SwapUsedPercent() float64 {
	if m.SwapTotalKB == 0 {
		return 0
	}

	return float64(m.SwapTotalKB-m.SwapAvailableKB) / float64(m.SwapTotalKB) * 100

}

// Read from /proc/meminfo
func ReadMem() (Mem, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return Mem{}, err
	}

	defer f.Close()

	var m Mem
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)

		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			m.TotalKB = value
		case "MemAvailable":
			m.AvailableKB = value
		case "SwapTotal":
			m.SwapTotalKB = value
		case "SwapFree":
			m.SwapAvailableKB = value
		}

	}
	return m, scanner.Err()
}
