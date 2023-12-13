package main

import (
	"flag"
	"os"
)

const ServerAddress = "SERVER_ADDRESS"

var addr string

// parseFlags parse startup options from cli params
func parseFlags() {
	flag.StringVar(&addr, "a", ":14444", "server address")
}

// parseEnv parse startup options from environment variables
func parseEnv() {
	if a := os.Getenv(ServerAddress); a != "" {
		addr = a
	}
}

// parseConfig Runs parse methods in priority env > flags
func parseConfig() {
	parseFlags()
	parseEnv()
}
