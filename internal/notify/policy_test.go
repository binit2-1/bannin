package notify

import (
	"testing"
	"time"

	"github.com/binit2-1/bannin/internal/engine"
)

func TestPolicy(t *testing.T) {
	p := DefaultPolicy()
	t0 := time.Unix(0, 0)
	at := func(s int) time.Time { return t0.Add(time.Duration(s) * time.Second) }

	steps := []struct {
		name     string
		old, new engine.Level
		t        time.Time
		want     bool
	}{
		{"elevated is quiet", engine.OK, engine.Elevated, at(0), false},
		{"high notifies", engine.Elevated, engine.High, at(1), true},
		{"critical gets through right after", engine.High, engine.Critical, at(2), true},
		{"going down is quiet", engine.Critical, engine.High, at(3), false},
		{"high again within cool-down", engine.Elevated, engine.High, at(30), false},
		{"critical again within cool-down", engine.High, engine.Critical, at(31), false},
		{"after cool-down", engine.Elevated, engine.High, at(200), true},
	}
	for _, s := range steps {
		if got := p.Should(s.old, s.new, s.t); got != s.want {
			t.Errorf("%s: got %v, want %v", s.name, got, s.want)
		}
	}

	p.Snooze(at(300), 10*time.Minute)
	if p.Should(engine.High, engine.Critical, at(400)) {
		t.Error("snoozed: should be quiet")
	}
	if !p.Should(engine.High, engine.Critical, at(300+601)) {
		t.Error("snooze over: should notify")
	}
}
