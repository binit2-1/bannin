package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/group"
	"github.com/binit2-1/bannin/internal/notify"
	"github.com/binit2-1/bannin/internal/sampler"
	"github.com/binit2-1/bannin/internal/stop"
)

// SnoozeFor is how long the Snooze button mutes warnings.
const SnoozeFor = 10 * time.Minute

// Alerts turns level changes into desktop notifications.
type Alerts struct {
	n    *notify.Notifier
	fake bool

	mu     sync.Mutex
	policy *notify.Policy
	target *group.Group // what the Stop button in the current popup stops
}

// NewAlerts connects to the notification server.
func NewAlerts(fake bool) (*Alerts, error) {
	n, err := notify.New()
	if err != nil {
		return nil, err
	}
	a := &Alerts{n: n, fake: fake, policy: notify.DefaultPolicy()}
	go a.handleActions()
	return a, nil
}

// Snooze mutes notifications for d.
func (a *Alerts) Snooze(d time.Duration) {
	a.mu.Lock()
	a.policy.Snooze(time.Now(), d)
	a.mu.Unlock()
	log.Printf("notifications snoozed for %v", d)
}

// SnoozedUntil is when the current snooze ends (zero or past if none).
func (a *Alerts) SnoozedUntil() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.policy.SnoozedUntil()
}

// OnLevel is hooked up to Service.OnLevel.
func (a *Alerts) OnLevel(old, new engine.Level, st Status) {
	a.mu.Lock()
	ok := a.policy.Should(old, new, time.Now())
	a.mu.Unlock()
	if !ok {
		return
	}

	target := stopCandidate()
	nt := buildNotification(new, st, target)

	a.mu.Lock()
	a.target = target
	a.mu.Unlock()

	if _, err := a.n.Show(nt); err != nil {
		log.Printf("notify: %v", err)
	}
}

// buildNotification writes the popup text. It is kept short: what is
// happening, how soon, and what the Stop button will do.
func buildNotification(l engine.Level, st Status, target *group.Group) notify.Notification {
	nt := notify.Notification{Icon: "dialog-warning", Urgency: notify.Normal}
	ram := fmt.Sprintf("RAM %.0f%%", st.Sample.MemPct)

	switch l {
	case engine.Critical:
		nt.Urgency = notify.Critical
		nt.Icon = "dialog-error"
		nt.Summary = "Memory almost full"
		if st.HasTTF {
			nt.Summary = fmt.Sprintf("Memory full in ~%ds", int(st.TTF.Seconds()))
		}
		nt.Body = ram + ". The system may freeze soon."
	default:
		nt.Summary = "Memory is getting high"
		nt.Body = ram + "."
	}

	if target != nil {
		nt.Body += fmt.Sprintf(" %s is using %.1f GB", target.Name, float64(target.RSSKB)/1024/1024)
		if target.Detached {
			nt.Body += " and was left running"
		}
		nt.Body += "."
		nt.Actions = append(nt.Actions, "stop", "Stop "+target.Name)
	}
	nt.Actions = append(nt.Actions, "snooze", "Snooze 10 min")
	return nt
}

// stopCandidate is the dev tool most worth stopping: a detached one first,
// then the biggest. Other apps (browsers, editors) are never suggested.
func stopCandidate() *group.Group {
	procs, err := sampler.ReadProcs()
	if err != nil {
		return nil
	}
	var best *group.Group
	for _, g := range group.Build(procs) { // biggest first
		if g.Kind == "other" {
			continue
		}
		if g.Detached {
			return &g
		}
		if best == nil {
			best = &g
		}
	}
	return best
}

func (a *Alerts) handleActions() {
	for act := range a.n.Actions() {
		switch act.Key {
		case "snooze":
			a.Snooze(SnoozeFor)
		case "stop":
			a.mu.Lock()
			t := a.target
			a.mu.Unlock()
			if t == nil {
				continue
			}
			if a.fake {
				log.Printf("(fake) would stop %s, PIDs %v", t.Name, t.PIDs)
				continue
			}
			alive := stop.Group(t.PIDs, false)
			log.Printf("stopped %s from notification (%d still running)", t.Name, len(alive))
		}
	}
}

// Close disconnects from D-Bus.
func (a *Alerts) Close() { a.n.Close() }
