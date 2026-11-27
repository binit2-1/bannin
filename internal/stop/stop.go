// Package stop ends processes safely.
//
// A PID can be reused: if Metro exits and the kernel gives its PID to a
// new process, a plain kill(pid) would hit the wrong one. So each target
// carries the start time we saw, and we open a pidfd (a handle to that one
// process), check the start time still matches, and send signals through
// the handle. Once the pidfd is open it can never point at another process.
package stop

import (
	"errors"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/unix"

	"github.com/binit2-1/bannin/internal/sampler"
)

// Target is one process to stop.
type Target struct {
	PID   int
	Start uint64 // start time from /proc/<pid>/stat; 0 skips the check
}

// handle is an open pidfd for a target we checked.
type handle struct {
	pid int
	fd  int
}

// Targets asks every process to exit with SIGTERM. With force=true,
// survivors get SIGKILL after 5 seconds. It returns the PIDs still alive.
// Processes that already exited, or whose PID now belongs to a different
// process, are skipped.
func Targets(targets []Target, force bool) []int {
	self := os.Getpid()

	var hs []handle
	for _, t := range targets {
		if t.PID == self || t.PID <= 1 {
			continue // never stop ourselves or init
		}
		fd, err := unix.PidfdOpen(t.PID, 0)
		if err != nil {
			continue // already gone
		}
		// Check after opening: from here on fd always means this process.
		if t.Start != 0 {
			if st, ok := sampler.StartTime(t.PID); !ok || st != t.Start {
				unix.Close(fd) // PID was reused, or it just exited
				continue
			}
		}
		hs = append(hs, handle{pid: t.PID, fd: fd})
	}
	defer func() {
		for _, h := range hs {
			unix.Close(h.fd)
		}
	}()

	for _, h := range hs {
		unix.PidfdSendSignal(h.fd, unix.SIGTERM, nil, 0) // polite request to exit
	}

	alive := waitExit(hs, 5*time.Second)
	if force && len(alive) > 0 {
		for _, h := range alive {
			unix.PidfdSendSignal(h.fd, unix.SIGKILL, nil, 0) // cannot be ignored
		}
		alive = waitExit(alive, time.Second)
	}

	pids := make([]int, 0, len(alive))
	for _, h := range alive {
		pids = append(pids, h.pid)
	}
	return pids
}

// waitExit polls until every process has exited or the timeout passes,
// and returns those still running.
func waitExit(hs []handle, timeout time.Duration) []handle {
	deadline := time.Now().Add(timeout)
	alive := hs
	for {
		alive = stillAlive(alive)
		if len(alive) == 0 || time.Now().After(deadline) {
			return alive
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func stillAlive(hs []handle) []handle {
	var out []handle
	for _, h := range hs {
		// Signal 0 checks without sending anything. ESRCH means it exited.
		// (A zombie still counts as alive until its parent reaps it, so also
		// treat state "Z" as gone: it holds no memory any more.)
		err := unix.PidfdSendSignal(h.fd, 0, nil, 0)
		if errors.Is(err, unix.ESRCH) || isZombie(h.pid) {
			continue
		}
		out = append(out, h)
	}
	return out
}

func isZombie(pid int) bool {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return true
	}
	s := string(b)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ')' {
			return i+2 < len(s) && s[i+2] == 'Z'
		}
	}
	return false
}
