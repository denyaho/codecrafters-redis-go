package main

import (
	"fmt"
	"net"
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit


func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	port := "6379"
	masterPort := ""
	masterAddr := ""
	role := "master"
	rdbdir := ""
	dbfilename := ""


	for i, arg := range os.Args {
		if arg == "--port" && i+1 < len(os.Args) {
			port = os.Args[i+1]
		}
		if arg == "--replicaof" && i+1 < len(os.Args) {
			role = "slave"
			masterInfo := strings.Split(os.Args[i + 1], " ")
			if len(masterInfo) != 2 {
				fmt.Printf("Invalid master address: %s\n", os.Args[i+1])
				return
			}
			masterAddr = masterInfo[0]
			masterPort = masterInfo[1]
		}
		if arg == "--dir" && i+1 <len(os.Args) {
			rdbdir = os.Args[i+1]
		}
		if arg == "--dbfilename" && i+1 <len(os.Args) {
			dbfilename = os.Args[i+1]
		}
	}
	rdb := rdb.NewRDB(rdbdir, dbfilename,)
	fmt.Println("Logs from your program will appear here!")
	server := server.New(port, "0.0.0.0", role, masterAddr + ":" + masterPort, rdb)
	server.StartServer()
	// Uncomment the code below to pass the first stage
	//
}
