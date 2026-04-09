package server

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func StartServer(port string, address string) {

	full_address := address + ":" + port
	l, err := net.Listen("tcp", full_address)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", port, err)
		os.Exit(1)
	}
	st := store.NewExpireMap()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handler.HandleConnection(conn, st)
	}
}