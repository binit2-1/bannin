package main

import (
	"fmt"
	"os"
)

const version = "0.0.1"

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("version: ", version)
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
