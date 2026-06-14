package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/aof"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
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
	currentdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rdbdir := currentdir
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
		if arg == "--appendonly" && i+1 <len(os.Args) {
			aofConfig.AppendOnly = os.Args[i+1]
		}
		if arg == "--appenddirname" && i+1 <len(os.Args) {
			aofConfig.AppendDirname = os.Args[i+1]
		}
		if arg == "--appendfilename" && i+1 <len(os.Args) {
			aofConfig.AppendFilename = os.Args[i+1]
		}
		if arg == "--appendfsync" && i+1 <len(os.Args) {
			aofConfig.AppendFsync = os.Args[i+1]
		}
	}
	rdb := rdb.NewRDB(rdbdir, dbfilename)
	aof := aof.NewAOF(rdbdir)
	st := store.NewExpireMap()
	if err := rdb.ReadFile(st); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("RDB file not found, starting with empty database")
		} else {
			fmt.Printf("Failed to load RDB file: %v\n", err)
			os.Exit(1)
		}
	}
	server := server.New(port, "0.0.0.0", role, masterAddr + ":" + masterPort, rdb, st, aof)
	server.StartServer()
	// Uncomment the code below to pass the first stage
	//
}
