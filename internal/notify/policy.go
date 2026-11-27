package notify

import (
	"time"

	"github.com/binit2-1/bannin/internal/engine"
)

// Policy decides when a level change is worth a notification, so the
// user gets one popup per real change and not a stream of them.
type Policy struct {
	MinLevel engine.Level  // quieter levels never notify (default High)
	Cooldown time.Duration // between notifications at the same level

	snoozedUntil time.Time
	lastAt       time.Time
	lastLevel    engine.Level
}

// DefaultPolicy notifies from High upwards, at most every 2 minutes per level.
func DefaultPolicy() *Policy {
	return &Policy{MinLevel: engine.High, Cooldown: 2 * time.Minute}
}

// Snooze mutes notifications until now+d.
func (p *Policy) Snooze(now time.Time, d time.Duration) { p.snoozedUntil = now.Add(d) }

// SnoozedUntil is when the snooze ends (zero if never snoozed).
func (p *Policy) SnoozedUntil() time.Time { return p.snoozedUntil }

// Should reports whether a change from old to new at time now should notify,
// and records it if so.
func (p *Policy) Should(old, new engine.Level, now time.Time) bool {
	if new <= old || new < p.MinLevel || now.Before(p.snoozedUntil) {
		return false
	}
	// A worse level than last time always gets through; the same level
	// again only after the cool-down (stops flapping High→Elevated→High).
	if new <= p.lastLevel && now.Sub(p.lastAt) < p.Cooldown {
		return false
	}
	p.lastAt, p.lastLevel = now, new
	return true
}
