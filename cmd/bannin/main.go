package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/binit2-1/bannin/internal/group"
	"github.com/binit2-1/bannin/internal/icon"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/sampler"
	"github.com/binit2-1/bannin/internal/stop"
)

const version = "0.0.1"

type memJSON struct {
	V       int     `json:"v"`
	MemPct  float64 `json:"mem_pct"`
	SwapPct float64 `json:"swap_pct"`
	UsedGB  float64 `json:"used_gb"`
	TotalGB float64 `json:"total_gb"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("version: ", version)
	case "mem":
		m, err := sampler.ReadMem()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		usedGB := float64(m.TotalKB-m.AvailableKB) / 1024 / 1024
		totalGB := float64(m.TotalKB) / 1024 / 1024

		if len(os.Args) > 2 && os.Args[2] == "--json" {
			out := ipc.MemStatus{
				V:       1,
				MemPct:  m.UsedPercent(),
				SwapPct: m.SwapUsedPercent(),
				UsedGB:  usedGB,
				TotalGB: totalGB,
			}

			json.NewEncoder(os.Stdout).Encode(out)
			return
		}
		fmt.Printf("RAM  %.1f%% used (%.1f of %.1f GB)\n",
			m.UsedPercent(),
			usedGB,
			totalGB,
		)
		fmt.Printf("Swap %.1f%% used\n", m.SwapUsedPercent())

	case "top":
		procs, err := sampler.ReadProcs()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].RSSKB > procs[j].RSSKB
		})
		if len(procs) > 15 {
			procs = procs[:15]
		}
		fmt.Printf("%7s %9s  %-16s %s\n", "PID", "RSS(MB)", "NAME", "COMMAND")
		for _, p := range procs {
			fmt.Printf("%7d %9.0f  %-16s %s\n",
				p.PID, float64(p.RSSKB)/1024, p.Name, truncate(p.Cmd, 70))
		}

	case "list":
		procs, err := sampler.ReadProcs()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		minMB := 100.0
		asJSON := false
		for _, a := range os.Args[2:] {
			switch a {
			case "--all":
				minMB = 0
			case "--json":
				asJSON = true
			}
		}

		icons := icon.New(icon.DefaultDirs()...)
		infos := make([]ipc.GroupInfo, 0) // empty list, not nil, so JSON gives [] not null
		for _, g := range group.Build(procs) {
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

		if asJSON {
			json.NewEncoder(os.Stdout).Encode(ipc.ListResponse{V: 1, Groups: infos})
			return
		}

		fmt.Printf("%-2s %-30s %9s %5s\n", "", "GROUP", "RSS(MB)", "PROCS")
		for _, g := range infos {
			mark := ""
			if g.Detached {
				mark = "⚠"
			}
			fmt.Printf("%-2s %-30s %9.0f %5d\n", mark, g.Name, g.RSSMB, len(g.PIDs))
		}
	case "stop":
		os.Exit(runStop(os.Args[2:]))

	default:
		fmt.Println("unknown command:", os.Args[1])
		usage()
		os.Exit(1)
	}

}

func usage() {
	fmt.Println("usage: bannin <command>")
	fmt.Println("commands:")
	fmt.Println("  version          show the version")
	fmt.Println("  mem [--json]     show RAM and swap use")
	fmt.Println("  top              biggest processes")
	fmt.Println("  list [--all] [--json]  processes grouped by app")
	fmt.Println("  stop <id> [--force]    stop a group from `list --json`")
}

// truncate shortens s to n characters, adding … if it was cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// runStop stops the group with the given ID and returns the exit code.
func runStop(args []string) int {
	id, force := "", false
	for _, a := range args {
		if a == "--force" {
			force = true
		} else {
			id = a
		}
	}
	if id == "" {
		fmt.Fprintln(os.Stderr, "usage: bannin stop <id> [--force]")
		return 2
	}

	procs, err := sampler.ReadProcs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	for _, g := range group.Build(procs) {
		if g.ID != id {
			continue
		}
		alive := stop.Group(g.PIDs, force)
		if len(alive) > 0 {
			fmt.Fprintf(os.Stderr, "%s: %d process(es) still running\n", g.Name, len(alive))
			return 1
		}
		fmt.Printf("stopped %s (%.0f MB)\n", g.Name, float64(g.RSSKB)/1024)
		return 0
	}
	fmt.Fprintln(os.Stderr, "no group with id", id)
	return 1
}
