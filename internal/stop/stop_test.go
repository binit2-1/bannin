package stop

import (
	"os/exec"
	"testing"

	"github.com/binit2-1/bannin/internal/sampler"
)

func start(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go cmd.Wait() // reap it so it does not stay a zombie
	return cmd
}

func TestStopsMatchingProcess(t *testing.T) {
	cmd := start(t)
	st, ok := sampler.StartTime(cmd.Process.Pid)
	if !ok {
		t.Fatal("no start time")
	}
	if alive := Targets([]Target{{PID: cmd.Process.Pid, Start: st}}, false); len(alive) != 0 {
		t.Fatalf("still alive: %v", alive)
	}
}

func TestSkipsRecycledPID(t *testing.T) {
	cmd := start(t)
	defer cmd.Process.Kill()
	st, _ := sampler.StartTime(cmd.Process.Pid)

	// Pretend we saw a different process with this PID earlier.
	Targets([]Target{{PID: cmd.Process.Pid, Start: st + 12345}}, true)

	if _, ok := sampler.StartTime(cmd.Process.Pid); !ok {
		t.Fatal("stopped a process whose start time did not match")
	}
}
