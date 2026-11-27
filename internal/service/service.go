// Package service is the bannind main loop: sample memory every second,
// keep a short history, and work out the warning level.
package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/binit2-1/bannin/internal/config"
	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/history"
	"github.com/binit2-1/bannin/internal/sampler"
)

// Options configure a Service.
type Options struct {
	Interval time.Duration
	Fake     bool // replay engine.FakeCurve instead of reading /proc
	Config   config.Config
}

// Status is a snapshot of what the service knows right now.
type Status struct {
	Sample   history.Sample
	Level    engine.Level
	TTF      time.Duration // time until RAM is full; 0 if unknown
	HasTTF   bool
	Total    uint64 // MemTotal in kB
	Fake     bool
	Received time.Time
}

// Service holds the state shared between the loop and its readers.
type Service struct {
	opts    Options
	started time.Time

	mu      sync.Mutex
	hist    *history.Ring
	tracker engine.Tracker
	status  Status

	Actions *ActionLog
	Auto    *AutoStopper

	// OnLevel is called (from the loop goroutine) when the level changes.
	OnLevel func(old, new engine.Level, st Status)
}

// New makes a service. Call Run to start it.
func New(opts Options) *Service {
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	actions := &ActionLog{}
	return &Service{
		opts:    opts,
		hist:    history.New(120),
		tracker: engine.Tracker{T: opts.Config.Thresholds},
		Actions: actions,
		Auto:    &AutoStopper{log: actions, fake: opts.Fake, cfg: opts.Config.AutoStop},
	}
}

// SetThresholds changes the level lines (used on reload).
func (s *Service) SetThresholds(t engine.Thresholds) {
	s.mu.Lock()
	s.tracker.T = t
	s.mu.Unlock()
}

// Run samples until ctx is cancelled.
func (s *Service) Run(ctx context.Context) error {
	s.started = time.Now()
	t := time.NewTicker(s.opts.Interval)
	defer t.Stop()

	s.tick(time.Now())
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-t.C:
			s.tick(now)
		}
	}
}

// Status returns the latest snapshot.
func (s *Service) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// History returns the samples from the last d, oldest first.
func (s *Service) History(d time.Duration) []history.Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hist.Since(time.Now().Add(-d))
}

func (s *Service) tick(now time.Time) {
	smp := history.Sample{T: now}
	var total uint64
	if s.opts.Fake {
		smp.MemPct, smp.PSI = engine.FakeCurve(now.Sub(s.started))
		total = 16 * 1024 * 1024
	} else {
		m, err := sampler.ReadMem()
		if err != nil {
			log.Printf("read meminfo: %v", err)
			return
		}
		smp.MemPct = m.UsedPercent()
		smp.SwapPct = m.SwapUsedPercent()
		total = m.TotalKB
		if p, err := engine.ReadPSI(); err == nil {
			smp.PSI = p.SomeAvg10
		}
	}

	s.mu.Lock()
	s.hist.Add(smp)
	old := s.tracker.Level()
	lvl, changed := s.tracker.Update(smp.MemPct, smp.PSI)
	ttf, hasTTF := engine.TimeToFull(s.hist.Since(now.Add(-30 * time.Second)))
	s.status = Status{
		Sample: smp, Level: lvl, TTF: ttf, HasTTF: hasTTF,
		Total: total, Fake: s.opts.Fake, Received: now,
	}
	st := s.status
	s.mu.Unlock()

	if changed {
		if hasTTF {
			log.Printf("level %v → %v (RAM %.1f%%, PSI %.1f, full in ~%ds)",
				old, lvl, smp.MemPct, smp.PSI, int(ttf.Seconds()))
		} else {
			log.Printf("level %v → %v (RAM %.1f%%, PSI %.1f)", old, lvl, smp.MemPct, smp.PSI)
		}
		if s.OnLevel != nil {
			s.OnLevel(old, lvl, st)
		}
	}
	s.Auto.Check(now, lvl)
}
