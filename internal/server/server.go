package server

import (
	"fmt"
	"net"
	"os"
	"math/rand"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

type Server struct {
	addr string
	st   *store.ExpireMap
	replicaManager *replication.ReplicaManager
}

func _generateReplID() string {
	randomstring := "abcdefghijklmnopqrstuvwxyz0123456789"
	replID := make([]byte, 40)
	for i := range replID {
		replID[i] = randomstring[rand.Intn(len(randomstring))]
	}
	return string(replID)
}

func New(port, address, role, masterAddr string) *Server {
	full_address := address + ":" + port
	return &Server{
		addr: full_address,
		st: store.NewExpireMap(),
		replicaManager: &replication.ReplicaManager{
			Role: role,
			ReplID: _generateReplID(),
			MasterAddr: masterAddr,
			Connections: []net.Conn{},
			IsPsynced: false,
		},
	}
}

func (s *Server) StartServer() {
	if s.replicaManager.Role == "slave" {
		go s.connectToMaster()
	}
	if s.replicaManager.Role == "master" {
		go s.replicaManager.StartTimer()
	}
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
			handler.HandleConnection(c, s.st, s.replicaManager)
		}(conn)
	}
}

func (s *Server) connectToMaster() {
	conn,err := net.Dial("tcp", s.replicaManager.MasterAddr)
	if err != nil {
		fmt.Printf("Failed to connect to master at %s: %v\n", s.replicaManager.MasterAddr, err)
		return
	}
	handler.HandleConnect_to_Master(conn,s.st, s.replicaManager)
	defer conn.Close()
}

