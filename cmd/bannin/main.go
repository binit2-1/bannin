package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/binit2-1/bannin/internal/sampler"
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
			out := memJSON{
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
		procs, err := sampler.ReadProc()
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
	default:
		fmt.Println("unknown command:", os.Args[1])
		usage()
		os.Exit(1)
	}

}

func usage() {
	fmt.Println("usage: bannin <command>")
	fmt.Println("commands:")
	fmt.Println("  version   show the version")
}

// truncate shortens s to n characters, adding … if it was cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
