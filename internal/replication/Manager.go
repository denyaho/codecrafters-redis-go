package replication

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ReplicaManager struct {
	Role string
	ReplID string
	MasterAddr string
	Connections []net.Conn
	mu          sync.Mutex
	IsPsynced bool
	Offset int64
}

func (rm *ReplicaManager) Add(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Connections = append(rm.Connections, conn)
}

func BuildCommand(args []string) []byte {
	responses := fmt.Sprintf("*%d\r\n",len(args))
	for _, arg := range args {
		responses += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	return []byte(responses)
}

func (rm *ReplicaManager) PropagateCommand(args []string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	command := BuildCommand(args)
	for _, conn := range rm.Connections {
		_, err := conn.Write(command)
		if err != nil {
			fmt.Printf("Failed to propagate command to replica: %v\n", err)
		}
	}
}

func (rm *ReplicaManager) SetPsynced(isPsynced bool){
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.IsPsynced = isPsynced
}

func (rm *ReplicaManager) GetPsynced() bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.IsPsynced
}

func (rm *ReplicaManager) AddOffset(args []string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	offset := int64(len(BuildCommand(args)))
	rm.Offset += offset
}

func (rm *ReplicaManager) GetOffset() int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.Offset
}

func (rm *ReplicaManager) SetOffset(){
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Offset = 0
}

func (rm *ReplicaManager) StartReplicationHeartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !rm.GetPsynced() {
			rm.PropagateCommand([]string{"REPLCONF","GETACK", "*"})
		}
	}

}
