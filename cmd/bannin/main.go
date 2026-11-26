package main

import (
	"fmt"
	"os"

	"github.com/binit2-1/bannin/internal/sampler"
)

const version = "0.0.1"

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

		fmt.Printf("RAM %.1f%% use (%.1f of %.1f GB)\n",
			m.UsedPercent(),
			float64(m.TotalKB-m.AvailableKB)/1024/1024,
			float64(m.TotalKB)/1024/1024,
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
