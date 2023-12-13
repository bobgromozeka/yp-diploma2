package main

import (
	"github.com/bobgromozeka/yp-diploma2/internal/server"
)

func main() {
	parseConfig()

	server.Run(addr)
}
