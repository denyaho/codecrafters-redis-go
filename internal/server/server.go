package server

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Server struct {
	addr string
	st   *store.ExpireMap
}

func New(port, address string) *Server {
	full_address := address + ":" + port
	return &Server{
		addr: full_address,
		st: store.NewExpireMap(),
	}
}

func (s *Server) StartServer() {

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", s.addr, err)
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go func(c net.Conn) {
			handler.HandleConnection(c, s.st)
		}(conn)
	}
}