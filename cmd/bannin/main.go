package main

import (
	"encoding/json"
	"fmt"
	"os"

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
