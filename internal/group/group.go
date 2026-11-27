package group

import (
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sort"
	"strings"

	"github.com/binit2-1/bannin/internal/sampler"
	"github.com/binit2-1/bannin/internal/stop"
)

// Group is a set of related processes shown as one line.
type Group struct {
	ID       string
	Name     string
	Kind     string // "metro", "gradle", "emulator", "devtool", "other"
	RSSKB    uint64
	PIDs     []int
	Starts   []uint64 // start time of each PID, same order
	Cmd      string   // command line of the biggest process
	topKB    uint64   // RSS of that process
	Detached bool     // a dev-tool process with no terminal and no real parent
}

// Build turns a flat process list into groups, biggest first.
func Build(procs []sampler.Proc) []Group {
	byPID := make(map[int]sampler.Proc, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}

	groups := map[string]*Group{}
	for _, p := range procs {
		kind, name := classify(p)

		g, ok := groups[name]
		if !ok {
			g = &Group{ID: makeID(name), Name: name, Kind: kind}
			groups[name] = g
		}
		g.RSSKB += p.RSSKB
		g.PIDs = append(g.PIDs, p.PID)
		g.Starts = append(g.Starts, p.Start)
		if p.RSSKB > g.topKB {
			g.topKB, g.Cmd = p.RSSKB, p.Cmd
		}

		if kind != "other" && isDetached(p, byPID) {
			g.Detached = true
		}
	}

	result := make([]Group, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].RSSKB > result[j].RSSKB
	})
	return result
}

// firefoxParts are the names Firefox gives its helper processes.
// (Linux cuts process names to 15 characters.)
var firefoxParts = map[string]bool{
	"firefox": true, "Isolated Web Co": true, "Web Content": true,
	"WebExtensions": true, "Privileged Cont": true, "RDD Process": true,
	"Socket Process": true, "Utility Process": true, "GPU Process": true,
	"Isolated Servic": true, "Forked Child": true,
}

// classify decides which group a process belongs to.
// The order matters: the first matching rule wins.
func classify(p sampler.Proc) (kind, name string) {
	cmd := strings.ToLower(p.Cmd)

	switch {
	case strings.Contains(cmd, "gradledaemon"),
		strings.Contains(cmd, "kotlincompiledaemon"):
		return "gradle", "Gradle daemons"

	case strings.HasPrefix(p.Name, "qemu-system"),
		strings.Contains(cmd, "/emulator/emulator"):
		return "emulator", "Android emulator"

	case p.Name == "node" && (strings.Contains(cmd, "metro") ||
		strings.Contains(cmd, "react-native") ||
		strings.Contains(cmd, "expo")):
		return "metro", "Metro · " + project(p)

	case firefoxParts[p.Name]:
		return "other", "Firefox"

	case p.Name == "adb", p.Name == "watchman":
		return "devtool", p.Name
	}
	return "other", p.Name
}

// project returns the name of the folder the process was started in.
func project(p sampler.Proc) string {
	if p.Cwd == "" || p.Cwd == "/" {
		return "unknown"
	}
	return filepath.Base(p.Cwd)
}

// isDetached: no terminal, and the parent is init or systemd.
// That is what a process looks like after its terminal was closed.
func isDetached(p sampler.Proc, byPID map[int]sampler.Proc) bool {
	if p.HasTTY {
		return false
	}
	if p.PPID == 1 {
		return true
	}
	parent, ok := byPID[p.PPID]
	return ok && parent.Name == "systemd"
}

// Targets returns the group's processes in the form stop.Targets wants.
func (g Group) Targets() []stop.Target {
	t := make([]stop.Target, len(g.PIDs))
	for i, pid := range g.PIDs {
		t[i] = stop.Target{PID: pid, Start: g.Starts[i]}
	}
	return t
}

// makeID turns a group name into a short ID made only of "g" and hex digits.
func makeID(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return fmt.Sprintf("g%08x", h.Sum32())
}
