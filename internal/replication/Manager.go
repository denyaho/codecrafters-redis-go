package replication

import (
	"fmt"
	"net"
	"sync"
	"time"
	"math/rand"
)

type ReplicaManager struct {
	Role string
	ReplID string
	MasterAddr string
	Connections []net.Conn
	mu          sync.Mutex
	IsPsynced bool
	Slaveoffset int64
	Masteroffset int64
	AckChan chan int64
}

func _generateReplID() string {
	randomstring := "abcdefghijklmnopqrstuvwxyz0123456789"
	replID := make([]byte, 40)
	for i := range replID {
		replID[i] = randomstring[rand.Intn(len(randomstring))]
	}
	return string(replID)
}

func NewReplicaManager(role, masterAddr string) *ReplicaManager {
	return &ReplicaManager{
		Role: role,
		ReplID: _generateReplID(),
		MasterAddr: masterAddr,
		Connections: []net.Conn{},
		AckChan: make(chan int64, 100),
	}
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
	rm.Masteroffset += int64(len(command))
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
	rm.Slaveoffset += int64(len(BuildCommand(args)))
}

func (rm *ReplicaManager) GetOffset() int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.Slaveoffset
}

func (rm *ReplicaManager) SetOffset(){
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Slaveoffset = 0
}

func (rm *ReplicaManager) StartReplicationHeartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if rm.GetPsynced() {
			rm.PropagateCommand([]string{"REPLCONF","GETACK", "*"})
		}
	}

}
