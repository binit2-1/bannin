package sampler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Proc struct {
	PID    int    //process ID
	PPID   int    // parent process id
	Name   string //process name
	RSSKB  uint64 //memory held by process in ram (kB)
	Cmd    string //full
	HasTTY bool   // attached to a terminal?
	Cwd    string // folder the process was started in
}

// ReadProcs lists the processes owned by the current user that use RAM.
func ReadProcs() ([]Proc, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	myUID := strconv.Itoa(os.Geteuid())
	var procs []Proc

	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		if p, ok := readProc(pid, myUID); ok {
			procs = append(procs, p)
		}
	}
	return procs, nil

}

// readProc reads one process. ok is false if we should skip it.
func readProc(pid int, myUID string) (Proc, bool) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return Proc{}, false //process exited
	}

	p := Proc{
		PID: pid,
	}

	mine := false

	for _, line := range strings.Split(string(data), "\n") {
		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			p.Name = val
		case "PPid":
			p.PPID, _ = strconv.Atoi(val)
		case "Uid":
			f := strings.Fields(val)
			mine = len(f) > 0 && f[0] == myUID
		case "VmRSS":
			f := strings.Fields(val) // "123456 kB"
			if len(f) > 0 {
				p.RSSKB, _ = strconv.ParseUint(f[0], 10, 64)
			}
		}
	}
	// Skip other users' processes, and kernel threads (they have no VmRSS).
	if !mine || p.RSSKB == 0 {
		return Proc{}, false
	}

	// Node 24 names its main thread "MainThread", and that shows up as the
	// process name. Use the name of the program file instead.
	if p.Name == "MainThread" {
		if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
			p.Name = filepath.Base(strings.TrimSuffix(exe, " (deleted)"))
		}
	}

	// cmdline separates arguments with NUL bytes (\x00).
	cmd, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	p.Cmd = strings.TrimSpace(strings.ReplaceAll(string(cmd), "\x00", " "))

	// /proc/<pid>/stat: the name is in (...) and can contain spaces or ")",
	// so we cut after the LAST ")". Then the fields are:
	// state, ppid, pgrp, session, tty_nr
	if stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		s := string(stat)
		if i := strings.LastIndex(s, ")"); i >= 0 {
			f := strings.Fields(s[i+1:])
			if len(f) > 4 {
				p.HasTTY = f[4] != "0"
			}
		}
	}

	p.Cwd, _ = os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))

	return p, true
}
