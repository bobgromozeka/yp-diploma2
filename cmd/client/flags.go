package main

import (
	"flag"
	"os"
)

const ServerAddress = "SERVER_ADDRESS"

var addr string

func parseFlags() {
	flag.StringVar(&addr, "a", ":14444", "server address")
}

func parseEnv() {
	if a := os.Getenv(ServerAddress); a != "" {
		addr = a
	}
}

func parseConfig() {
	parseFlags()
	parseEnv()
}
