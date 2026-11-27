package history

import (
	"testing"
	"time"
)

func TestRing(t *testing.T) {
	r := New(3)
	if _, ok := r.Last(); ok {
		t.Fatal("empty ring has a last sample")
	}
	t0 := time.Unix(1000, 0)
	for i := range 5 {
		r.Add(Sample{T: t0.Add(time.Duration(i) * time.Second), MemPct: float64(i)})
	}
	if r.Len() != 3 {
		t.Fatalf("Len = %d, want 3", r.Len())
	}
	all := r.All()
	for i, want := range []float64{2, 3, 4} {
		if all[i].MemPct != want {
			t.Errorf("All()[%d] = %v, want %v", i, all[i].MemPct, want)
		}
	}
	if last, _ := r.Last(); last.MemPct != 4 {
		t.Errorf("Last = %v, want 4", last.MemPct)
	}
	if got := r.Since(t0.Add(3 * time.Second)); len(got) != 2 {
		t.Errorf("Since gave %d samples, want 2", len(got))
	}
}
