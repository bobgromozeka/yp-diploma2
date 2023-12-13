package main

import (
	"fmt"
	"os"
)

var (
	Version   string
	BuildTime string
)

// build shows current build information
func build() {
	if len(os.Args) > 1 && (os.Args[1] == "--v" || os.Args[1] == "--version") {
		fmt.Printf("version=%s, time=%s\n", Version, BuildTime)
		os.Exit(0)
	}
}
