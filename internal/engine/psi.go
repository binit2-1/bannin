package engine

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

// PSI is memory pressure: the share of time (0-100) that tasks were
// stalled waiting for memory, averaged over the last 10 seconds.
// "Some" means at least one task was stalled; "Full" means all were.
type PSI struct {
	SomeAvg10 float64
	FullAvg10 float64
}

// ReadPSI reads /proc/pressure/memory. Kernels without PSI return an error,
// and callers should then treat pressure as zero.
func ReadPSI() (PSI, error) {
	f, err := os.Open("/proc/pressure/memory")
	if err != nil {
		return PSI{}, err
	}
	defer f.Close()
	return parsePSI(f)
}

// parsePSI reads lines like:
//
//	some avg10=1.23 avg60=0.50 avg300=0.10 total=12345
func parsePSI(r io.Reader) (PSI, error) {
	var p PSI
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		val, ok := strings.CutPrefix(fields[1], "avg10=")
		if !ok {
			continue
		}
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "some":
			p.SomeAvg10 = v
		case "full":
			p.FullAvg10 = v
		}
	}
	return p, sc.Err()
}
