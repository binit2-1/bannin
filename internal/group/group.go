package group

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/binit2-1/bannin/internal/sampler"
)

// Group is a set of related processes shown as one line.
type Group struct {
	Name     string
	Kind     string // "metro", "gradle", "emulator", "devtool", "other"
	RSSKB    uint64
	PIDs     []int
	Detached bool // a dev-tool process with no terminal and no real parent
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
			g = &Group{Name: name, Kind: kind}
			groups[name] = g
		}
		g.RSSKB += p.RSSKB
		g.PIDs = append(g.PIDs, p.PID)

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
