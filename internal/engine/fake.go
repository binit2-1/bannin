package engine

import "time"

// FakeCurve is a made-up memory curve for testing, so the levels and
// notifications can be seen without really filling RAM. One cycle is
// 2 minutes: it climbs from 50 % to 97 %, holds, then falls back.
func FakeCurve(elapsed time.Duration) (memPct, psi float64) {
	s := elapsed.Seconds()
	s -= float64(int(s/120) * 120) // repeat every 120 s

	switch {
	case s < 20: // calm
		memPct = 50 + s
	case s < 60: // something leaks: 70 → 97
		memPct = 70 + (s-20)*27/40
	case s < 75: // stuck near full, the system starts stalling
		memPct = 97
	case s < 105: // something was stopped: back down
		memPct = 97 - (s-75)*42/30
	default:
		memPct = 55
	}
	if memPct > 90 {
		psi = (memPct - 90) * 4 // up to ~28
	}
	return memPct, psi
}
