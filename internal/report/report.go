// Package report builds the status and process list that the CLI prints
// and the service sends over its socket.
package report

import (
	"fmt"

	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/group"
	"github.com/binit2-1/bannin/internal/icon"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/sampler"
	"github.com/binit2-1/bannin/internal/stop"
)

// Groups reads the process list and groups it, biggest first.
func Groups() ([]group.Group, error) {
	procs, err := sampler.ReadProcs()
	if err != nil {
		return nil, err
	}
	return group.Build(procs), nil
}

// Status reads memory and pressure directly, with the level for this
// single sample (no history, so no time-to-full).
func Status(th engine.Thresholds) (ipc.MemStatus, error) {
	m, err := sampler.ReadMem()
	if err != nil {
		return ipc.MemStatus{}, err
	}
	st := ipc.MemStatus{
		V:       1,
		MemPct:  m.UsedPercent(),
		SwapPct: m.SwapUsedPercent(),
		UsedGB:  float64(m.TotalKB-m.AvailableKB) / 1024 / 1024,
		TotalGB: float64(m.TotalKB) / 1024 / 1024,
	}
	if p, err := engine.ReadPSI(); err == nil {
		st.PSI = p.SomeAvg10
	}
	st.Level = th.Classify(st.MemPct, st.PSI).String()
	if groups, err := Groups(); err == nil {
		AddGroupSummary(&st, groups)
	}
	return st, nil
}

// AddGroupSummary fills in Warn and Biggest.
func AddGroupSummary(st *ipc.MemStatus, groups []group.Group) {
	st.Warn, st.Biggest = 0, ""
	if len(groups) > 0 {
		st.Biggest = groups[0].Name
	}
	for _, g := range groups {
		if g.Detached {
			st.Warn++
		}
	}
}

// List turns groups into the JSON list, skipping those under minMB.
func List(groups []group.Group, minMB float64, icons *icon.Resolver) ipc.ListResponse {
	infos := make([]ipc.GroupInfo, 0) // empty list, not nil, so JSON gives [] not null
	for _, g := range groups {
		mb := float64(g.RSSKB) / 1024
		if mb < minMB {
			continue
		}
		infos = append(infos, ipc.GroupInfo{
			ID:       g.ID,
			Name:     g.Name,
			Kind:     g.Kind,
			RSSMB:    mb,
			PIDs:     g.PIDs,
			Detached: g.Detached,
			Icon:     icons.For(g.Kind, g.Name),
		})
	}
	return ipc.ListResponse{V: 1, Groups: infos}
}

// Stop stops the group with the given ID.
func Stop(id string, force bool) ipc.Result {
	groups, err := Groups()
	if err != nil {
		return ipc.Result{Error: err.Error()}
	}
	for _, g := range groups {
		if g.ID != id {
			continue
		}
		if alive := stop.Targets(g.Targets(), force); len(alive) > 0 {
			return ipc.Result{Error: fmt.Sprintf("%s: %d process(es) still running", g.Name, len(alive))}
		}
		return ipc.Result{OK: true, Message: fmt.Sprintf("stopped %s (%.0f MB)", g.Name, float64(g.RSSKB)/1024)}
	}
	return ipc.Result{Error: "no group with id " + id}
}
