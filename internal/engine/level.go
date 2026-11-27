// Package engine turns memory samples into a warning level.
package engine

// Level is how worried Bannin is about memory.
type Level int

const (
	OK Level = iota
	Elevated
	High
	Critical
)

func (l Level) String() string {
	switch l {
	case Elevated:
		return "elevated"
	case High:
		return "high"
	case Critical:
		return "critical"
	}
	return "ok"
}

// Thresholds are the RAM percentages (and pressure values) where each level starts.
type Thresholds struct {
	Elevated float64 `toml:"elevated"`
	High     float64 `toml:"high"`
	Critical float64 `toml:"critical"`

	// Memory pressure (PSI "some" avg10) can raise the level on its own:
	// a machine that is swapping hard is in trouble even at 80 % RAM.
	PSIHigh     float64 `toml:"psi_high"`
	PSICritical float64 `toml:"psi_critical"`

	// Hysteresis: to go down a level, RAM must drop this many points
	// below the threshold. This stops the level flapping around a line.
	Hysteresis float64 `toml:"hysteresis"`
}

// DefaultThresholds are used when the config file does not set any.
var DefaultThresholds = Thresholds{
	Elevated:    75,
	High:        85,
	Critical:    92,
	PSIHigh:     10,
	PSICritical: 25,
	Hysteresis:  3,
}

// Classify returns the level for one sample, with no memory of the past.
func (t Thresholds) Classify(memPct, psi float64) Level {
	l := OK
	switch {
	case memPct >= t.Critical:
		l = Critical
	case memPct >= t.High:
		l = High
	case memPct >= t.Elevated:
		l = Elevated
	}
	switch {
	case psi >= t.PSICritical && l < Critical:
		l = Critical
	case psi >= t.PSIHigh && l < High:
		l = High
	}
	return l
}

// Tracker remembers the current level and applies hysteresis.
type Tracker struct {
	T     Thresholds
	level Level
}

// Level is the current level.
func (tr *Tracker) Level() Level { return tr.level }

// Update feeds in a sample and returns the new level, and whether it changed.
// The level goes up at once, but only goes down once RAM (and pressure)
// are clearly below the line.
func (tr *Tracker) Update(memPct, psi float64) (Level, bool) {
	old := tr.level
	if raw := tr.T.Classify(memPct, psi); raw > tr.level {
		tr.level = raw
	} else {
		// Pretend the sample is a bit worse than it is; if even that
		// is a lower level, we have really calmed down.
		h := tr.T.Hysteresis
		if lower := tr.T.Classify(memPct+h, psi+h); lower < tr.level {
			tr.level = lower
		}
	}
	return tr.level, tr.level != old
}
