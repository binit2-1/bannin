package stop

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Group asks every process in pids to exit.
// With force=true, survivors are killed after 5 seconds.
// It returns the PIDs that are still alive.
func Group(pids []int, force bool) []int {
	self := os.Getpid()

	var targets []int
	for _, pid := range pids {
		if pid == self {
			continue // never stop ourselves
		}
		targets = append(targets, pid)
		syscall.Kill(pid, syscall.SIGTERM) // polite request to exit
	}

	// Wait up to 5 seconds for them to exit.
	alive := targets
	deadline := time.Now().Add(5 * time.Second)
	for len(alive) > 0 && time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		alive = stillAlive(alive)
	}

	if force && len(alive) > 0 {
		for _, pid := range alive {
			syscall.Kill(pid, syscall.SIGKILL) // cannot be ignored
		}
		time.Sleep(300 * time.Millisecond)
		alive = stillAlive(alive)
	}
	return alive
}

func stillAlive(pids []int) []int {
	var alive []int
	for _, pid := range pids {
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err == nil {
			alive = append(alive, pid)
		}
	}
	return alive
}
