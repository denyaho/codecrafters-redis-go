package main

import (
	"net"
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	config := parseArgs()
	st, rdb, aof := initStorage(&config)

	server := server.New(config.Port, "0.0.0.0", config.Role, config.MasterAddr + ":" + config.MasterPort, rdb, st, aof)
	server.StartServer()
	// Uncomment the code below to pass the first stage
	//
}