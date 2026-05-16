package replication

import (
	"fmt"
	"net"
	"sync"
)

type ReplicaManager struct {
	Role string
	ReplID string
	MasterAddr string
	Connections []net.Conn
	mu          sync.Mutex
}

func (rm *ReplicaManager) Add(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Connections = append(rm.Connections, conn)
}

func _buildCommand(args []string) []byte {
	responses := fmt.Sprintf("*%d\r\n",len(args))
	for _, arg := range args {
		responses += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	return []byte(responses)
}

func (rm *ReplicaManager) PropagateCommand(args []string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	command := _buildCommand(args)
	for _, conn := range rm.Connections {
		_, err := conn.Write(command)
		if err != nil {
			fmt.Printf("Failed to propagate command to replica: %v\n", err)
		}
	}
}
