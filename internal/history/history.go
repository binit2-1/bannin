// Package history keeps the last few minutes of memory samples.
package history

import "time"

// Sample is one reading of the system.
type Sample struct {
	T       time.Time
	MemPct  float64
	SwapPct float64
	PSI     float64 // memory "some" avg10 from /proc/pressure/memory
}

// Ring is a fixed-size buffer: when it is full, the oldest sample is dropped.
// It is not safe for use from several goroutines at once.
type Ring struct {
	buf   []Sample
	start int // index of the oldest sample
	n     int // number of samples stored
}

// New makes a ring that holds up to size samples.
func New(size int) *Ring {
	if size < 1 {
		size = 1
	}
	return &Ring{buf: make([]Sample, size)}
}

// Add stores s, dropping the oldest sample if the ring is full.
func (r *Ring) Add(s Sample) {
	if r.n < len(r.buf) {
		r.buf[(r.start+r.n)%len(r.buf)] = s
		r.n++
		return
	}
	r.buf[r.start] = s
	r.start = (r.start + 1) % len(r.buf)
}

// Len is the number of samples stored.
func (r *Ring) Len() int { return r.n }

// Last returns the newest sample, and false if the ring is empty.
func (r *Ring) Last() (Sample, bool) {
	if r.n == 0 {
		return Sample{}, false
	}
	return r.buf[(r.start+r.n-1)%len(r.buf)], true
}

// Since returns the samples taken at or after t, oldest first, as a new slice.
func (r *Ring) Since(t time.Time) []Sample {
	out := make([]Sample, 0, r.n)
	for i := range r.n {
		s := r.buf[(r.start+i)%len(r.buf)]
		if !s.T.Before(t) {
			out = append(out, s)
		}
	}
	return out
}

// All returns every sample, oldest first, as a new slice.
func (r *Ring) All() []Sample { return r.Since(time.Time{}) }
