package engine

import (
	"time"

	"github.com/binit2-1/bannin/internal/history"
)

// TimeToFull guesses how long until RAM reaches 100 %, by fitting a
// straight line through the samples (least squares). ok is false when
// memory is not rising, there are too few samples, or the answer is
// more than 10 minutes away (too far out to be worth showing).
func TimeToFull(samples []history.Sample) (d time.Duration, ok bool) {
	if len(samples) < 5 {
		return 0, false
	}
	t0 := samples[0].T
	var n, sx, sy, sxx, sxy float64
	for _, s := range samples {
		x := s.T.Sub(t0).Seconds()
		y := s.MemPct
		n++
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
	}
	den := n*sxx - sx*sx
	if den == 0 {
		return 0, false
	}
	slope := (n*sxy - sx*sy) / den // percentage points per second
	if slope < 0.02 {
		return 0, false
	}
	intercept := (sy - slope*sx) / n
	lastX := samples[len(samples)-1].T.Sub(t0).Seconds()
	now := intercept + slope*lastX
	secs := (100 - now) / slope
	if secs < 0 {
		secs = 0
	}
	if secs > 600 {
		return 0, false
	}
	return time.Duration(secs * float64(time.Second)), true
}
