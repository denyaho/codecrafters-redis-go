package handler

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)


func handleInfo(st *store.ExpireMap, args []string, role, replID string) []byte {
	fields := [][2]string{
		{"role", role},
		{"master_replid", replID},
		{"connected_slaves", "0"},
		{"master_repl_offset", "0"},
		{"second_repl_offset", "-1"},
		{"repl_backlog_active", "0"},
		{"repl_backlog_size", "1048576"},
		{"repl_backlog_first_byte_offset", "0"},
		{"repl_backlog_histlen", "0"},
	}

	body := ""
	for _, kv := range fields {
		body += fmt.Sprintf("%s:%s\r\n", kv[0], kv[1])
	}
	response := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(body), body))
	return response
}

func handleREPLCONF(st *store.ExpireMap, args []string, rm *replication.ReplicaManager) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'REPLCONF' command\r\n")
	}
	for i := 1; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return []byte("-ERR wrong number of arguments for 'REPLCONF' command\r\n")
		}
		if args[i] == "listening-port" {
			return []byte("+OK\r\n")
		}
		if args[i] == "capa" {
			return []byte("+OK\r\n")
		}
		if args[i] == "ACK" {
			offset := args[i+1]
			offsetInt, err := strconv.ParseInt(offset, 10, 64)
			if err != nil {
				return []byte("-ERR invalid offset for 'ACK' in 'REPLCONF' command\r\n")
			}
			rm.AckChan <- offsetInt
			return []byte("+OK\r\n")
		}
	}
	return []byte("-ERR unknown REPLCONF option\r\n")
}

var RDBcontent, _ = hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")

func buildRDB() []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s", len(RDBcontent), RDBcontent))
}

func handlePSYNC(st *store.ExpireMap, args []string, replID string) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'PSYNC' command\r\n")
	}
	offset := args[2]
	if offset == "-1" {
		offset = "0"
	}
	if args[1] == "?" {
		return []byte(fmt.Sprintf("+FULLRESYNC %s %s\r\n", replID, offset))
	}
	return []byte(fmt.Sprintf("+FULLRESYNC %s %s\r\n", replID, offset))	
}

func handleWAIT(args []string, rm *replication.ReplicaManager) []byte {
	if len(args) != 3 {
		return []byte("-ERR wrong number of arguments for 'WAIT' command\r\n")
	}
	numreplicas, err := strconv.Atoi(args[1])
	if err != nil {
		return []byte("-ERR invalid number of replicas\r\n")
	}
	timeout, err := strconv.Atoi(args[2])
	if err != nil {
		return []byte("-ERR invalid timeout\r\n")
	}
	 if numreplicas == 0 || len(rm.Connections) == 0 {
      return []byte(fmt.Sprintf(":%d\r\n", len(rm.Connections)))}
	
	var timer <- chan time.Time
	if timeout > 0 {
		timer = time.After(time.Duration(timeout) * time.Millisecond)
	}
	if timeout == 0 {
		return []byte(fmt.Sprintf(":%d\r\n", len(rm.Connections)))
	}

	acked := 0
	rm.PropagateCommand([]string{"REPLCONF", "GETACK", "*"})
	for {
		select {
		case offset := <-rm.AckChan:
			if offset >= rm.Masteroffset {
				acked ++
				if acked >= numreplicas {
					return []byte(fmt.Sprintf(":%d\r\n", acked))
				}
			}
		case <-timer:
			return []byte(fmt.Sprintf(":%d\r\n", acked))
		}
	}	
}
