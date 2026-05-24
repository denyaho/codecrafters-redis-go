package handler

import (
	"fmt"
	"strconv"
	"time"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
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
	return resp.BuildBulkStrings(body)
}

func handleREPLCONF(st *store.ExpireMap, args []string, rm *replication.ReplicaManager) []byte {
	if len(args) < 3 {
		return resp.BuildError("ERR wrong number of arguments for 'REPLCONF' command")

	}
	for i := 1; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return resp.BuildError("ERR wrong number of arguments for 'REPLCONF' command")
		}
		if args[i] == "listening-port" {
			return resp.BuildSimpleString("OK")
		}
		if args[i] == "capa" {
			return resp.BuildSimpleString("OK")
		}
		if args[i] == "ACK" {
			offset := args[i+1]
			offsetInt, err := strconv.ParseInt(offset, 10, 64)
			if err != nil {
				return resp.BuildError("ERR invalid offset for 'ACK' in 'REPLCONF' command")
			}
			rm.AckChan <- offsetInt
			return nil
		}
	}

	return resp.BuildError("ERR unknown REPLCONF option")
}



func handlePSYNC(st *store.ExpireMap, args []string, replID string) []byte {
	if len(args) < 3 {
		return resp.BuildError("ERR wrong number of arguments for 'PSYNC' command")
	}
	offset := args[2]
	if offset == "-1" {
		offset = "0"
	}
	if args[1] == "?" {
		return resp.BuildSimpleString(fmt.Sprintf("FULLRESYNC %s %s", replID, offset))
	}
	return resp.BuildSimpleString(fmt.Sprintf("FULLRESYNC %s %s", replID, offset))	
}

func handleWAIT(args []string, rm *replication.ReplicaManager) []byte {
	if len(args) != 3 {
		return resp.BuildError("ERR wrong number of arguments for 'WAIT' command")
	}
	numreplicas, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.BuildError("ERR invalid number of replicas")
	}
	timeout, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.BuildError("ERR invalid timeout")
	}

	if rm.Masteroffset == 0 {
		return resp.BuildInteger(len(rm.Connections))
	}
	
	var timer <- chan time.Time
	if timeout > 0 {
		timer = time.After(time.Duration(timeout) * time.Millisecond)
	}
	if timeout == 0 {
		return resp.BuildInteger(len(rm.Connections))
	}

	acked := 0
	rm.PropagateCommand([]string{"REPLCONF", "GETACK", "*"})
	for {
		select {
		case <-rm.AckChan:
			acked++
			if acked >= numreplicas {
				return resp.BuildInteger(acked)
			}
		case <-timer:
			return resp.BuildInteger(acked)
		}
	}	
}
