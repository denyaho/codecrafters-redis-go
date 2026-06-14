package server

import (
	"fmt"
	"net"
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
	"github.com/codecrafters-io/redis-starter-go/internal/aof"
)

type Server struct {
	addr string
	st   *store.ExpireMap
	replicaManager *replication.ReplicaManager
	rdbConfig *rdb.RDB
	pubsubManager *pubsub.Manager
	aofConfig *aof.AOF
}

func New(port, address, role, masterAddr string, rdbconfig *rdb.RDB, st *store.ExpireMap, aofConfig *aof.AOF) *Server {
	full_address := address + ":" + port
	return &Server{
		addr: full_address,
		st: st,
		replicaManager: replication.NewReplicaManager(role, masterAddr),
		rdbConfig: rdbconfig,
		pubsubManager: pubsub.NewManager(),
		aofConfig: aofConfig,
	}
}

func (s *Server) StartServer() {
	if s.replicaManager.Role == "slave" {
		go s.connectToMaster()
	}
	if s.replicaManager.Role == "master" {
		go s.replicaManager.StartReplicationHeartbeat()
	}
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", s.addr, err)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		c := pubsub.NewClient(conn, s.pubsubManager)
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go func(c *pubsub.Client) {
			handler.HandleConnection(c, s.st, s.replicaManager, s.rdbConfig, s.pubsubManager, s.aofConfig)
		}(c)
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

