package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/binit2-1/bannin/internal/config"
	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/report"
	"github.com/binit2-1/bannin/internal/rules"
	"github.com/binit2-1/bannin/internal/stop"
)

// gap between two auto-stops, so memory can settle and be measured again.
const autoStopGap = 20 * time.Second

// AutoStopper stops (or, in dry run, names) dev tools when RAM stays critical.
type AutoStopper struct {
	log  *ActionLog
	fake bool

	mu       sync.Mutex
	cfg      config.AutoStop
	since    time.Time // when the current critical spell began
	lastAct  time.Time
	reported map[string]bool // dry run: groups already named this spell
}

// SetConfig replaces the settings (used on reload).
func (a *AutoStopper) SetConfig(c config.AutoStop) {
	a.mu.Lock()
	a.cfg = c
	a.mu.Unlock()
}

// Config returns the current settings.
func (a *AutoStopper) Config() config.AutoStop {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

// Check is called every tick with the current level.
func (a *AutoStopper) Check(now time.Time, lvl engine.Level) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if lvl < engine.Critical {
		a.since, a.reported = time.Time{}, nil
		return
	}
	if a.since.IsZero() {
		a.since, a.reported = now, map[string]bool{}
	}
	if !a.cfg.Enabled ||
		now.Sub(a.since) < time.Duration(a.cfg.AfterSeconds)*time.Second ||
		now.Sub(a.lastAct) < autoStopGap {
		return
	}

	groups, err := report.Groups()
	if err != nil {
		return
	}
	dry := a.cfg.DryRun || a.fake // the fake curve must never stop real processes
	for _, g := range rules.Pick(groups, a.cfg) {
		if dry && a.reported[g.ID] {
			continue
		}
		gb := float64(g.RSSKB) / 1024 / 1024
		a.lastAct = now
		if dry {
			a.reported[g.ID] = true
			a.log.Add(fmt.Sprintf("Would have stopped %s (%.1f GB)", g.Name, gb), true)
			return
		}
		alive := stop.Targets(g.Targets(), true)
		text := fmt.Sprintf("Stopped %s (%.1f GB)", g.Name, gb)
		if len(alive) > 0 {
			text += fmt.Sprintf(", %d process(es) would not exit", len(alive))
		}
		a.log.Add(text, false)
		return // one at a time, then measure again
	}
}
