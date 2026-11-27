package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/binit2-1/bannin/internal/history"
)

func TestParsePSI(t *testing.T) {
	in := "some avg10=12.50 avg60=3.00 avg300=1.00 total=6248427\n" +
		"full avg10=4.25 avg60=0.00 avg300=0.00 total=5651734\n"
	p, err := parsePSI(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if p.SomeAvg10 != 12.5 || p.FullAvg10 != 4.25 {
		t.Errorf("got %+v", p)
	}
}

func TestClassify(t *testing.T) {
	th := DefaultThresholds
	cases := []struct {
		mem, psi float64
		want     Level
	}{
		{50, 0, OK},
		{75, 0, Elevated},
		{86, 0, High},
		{95, 0, Critical},
		{60, 12, High},     // pressure alone
		{60, 30, Critical}, // heavy pressure alone
	}
	for _, c := range cases {
		if got := th.Classify(c.mem, c.psi); got != c.want {
			t.Errorf("Classify(%v, %v) = %v, want %v", c.mem, c.psi, got, c.want)
		}
	}
}

func TestHysteresis(t *testing.T) {
	tr := Tracker{T: DefaultThresholds}
	steps := []struct {
		mem     float64
		want    Level
		changed bool
	}{
		{80, Elevated, true},
		{86, High, true},
		{84, High, false},   // just below 85: stay High
		{82.5, High, false}, // 82.5+3 = 85.5, still High
		{81.9, Elevated, true},
		{50, OK, true},
	}
	for i, s := range steps {
		got, changed := tr.Update(s.mem, 0)
		if got != s.want || changed != s.changed {
			t.Errorf("step %d (%v%%): got %v/%v, want %v/%v", i, s.mem, got, changed, s.want, s.changed)
		}
	}
}

func TestTimeToFull(t *testing.T) {
	t0 := time.Unix(0, 0)
	var rising, flat []history.Sample
	for i := range 30 {
		ts := t0.Add(time.Duration(i) * time.Second)
		rising = append(rising, history.Sample{T: ts, MemPct: 70 + float64(i)*0.5})
		flat = append(flat, history.Sample{T: ts, MemPct: 70})
	}
	// Last point is 84.5 %, rising 0.5 points/s: 31 s to 100 %.
	d, ok := TimeToFull(rising)
	if !ok || d < 30*time.Second || d > 32*time.Second {
		t.Errorf("rising: got %v %v, want ~31s", d, ok)
	}
	if _, ok := TimeToFull(flat); ok {
		t.Error("flat: expected no estimate")
	}
	if _, ok := TimeToFull(rising[:3]); ok {
		t.Error("too few samples: expected no estimate")
	}
}

func TestFakeCurveReachesCritical(t *testing.T) {
	tr := Tracker{T: DefaultThresholds}
	seen := map[Level]bool{}
	for s := range 120 {
		m, p := FakeCurve(time.Duration(s) * time.Second)
		l, _ := tr.Update(m, p)
		seen[l] = true
	}
	for _, l := range []Level{OK, Elevated, High, Critical} {
		if !seen[l] {
			t.Errorf("fake curve never reached %v", l)
		}
	}
}
