package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/icon"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/report"
	"github.com/binit2-1/bannin/internal/sampler"
)

var version = "dev" // set by -ldflags at build time

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("version: ", version)
	case "mem":
		st, err := getStatus()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if len(os.Args) > 2 && os.Args[2] == "--json" {
			json.NewEncoder(os.Stdout).Encode(st)
			return
		}
		fmt.Printf("RAM  %.1f%% used (%.1f of %.1f GB)  level: %s\n",
			st.MemPct, st.UsedGB, st.TotalGB, st.Level)
		fmt.Printf("Swap %.1f%% used\n", st.SwapPct)
		if st.TTFSec > 0 {
			fmt.Printf("Full in about %ds\n", st.TTFSec)
		}

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
		all, asJSON := false, false
		for _, a := range os.Args[2:] {
			switch a {
			case "--all":
				all = true
			case "--json":
				asJSON = true
			}
		}
		list, err := getList(all)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if asJSON {
			json.NewEncoder(os.Stdout).Encode(list)
			return
		}
		fmt.Printf("%-2s %-30s %9s %5s\n", "", "GROUP", "RSS(MB)", "PROCS")
		for _, g := range list.Groups {
			mark := ""
			if g.Detached {
				mark = "⚠"
			}
			fmt.Printf("%-2s %-30s %9.0f %5d\n", mark, g.Name, g.RSSMB, len(g.PIDs))
		}

	case "snooze":
		mins := 10
		if len(os.Args) > 2 {
			fmt.Sscan(os.Args[2], &mins)
		}
		var res ipc.Result
		if err := ipc.Call(ipc.Request{Method: "snooze", Minutes: mins}, &res); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(res.Message)

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
	fmt.Println("  snooze [minutes]       mute warnings (needs bannind)")
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
// The service does it when running; otherwise we do it ourselves.
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

	var res ipc.Result
	err := ipc.Call(ipc.Request{Method: "stop", ID: id, Force: force}, &res)
	if errors.Is(err, ipc.ErrNoService) {
		res, err = report.Stop(id, force), nil
	}
	if err == nil && res.Error != "" {
		err = errors.New(res.Error)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Println(res.Message)
	return 0
}

// getStatus asks bannind, or reads /proc directly if it is not running.
func getStatus() (ipc.MemStatus, error) {
	var st ipc.MemStatus
	err := ipc.Call(ipc.Request{Method: "status"}, &st)
	if errors.Is(err, ipc.ErrNoService) {
		return report.Status(engine.DefaultThresholds)
	}
	return st, err
}

// getList asks bannind, or reads /proc directly if it is not running.
func getList(all bool) (ipc.ListResponse, error) {
	var list ipc.ListResponse
	err := ipc.Call(ipc.Request{Method: "list", All: all}, &list)
	if !errors.Is(err, ipc.ErrNoService) {
		return list, err
	}
	groups, err := report.Groups()
	if err != nil {
		return list, err
	}
	minMB := 100.0
	if all {
		minMB = 0
	}
	return report.List(groups, minMB, icon.New(icon.DefaultDirs()...)), nil
}
