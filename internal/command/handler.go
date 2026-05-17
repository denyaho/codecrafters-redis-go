package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)


func HandleConnection(conn net.Conn, st *store.ExpireMap, replicaManager *replication.ReplicaManager) {
	defer conn.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	reader := bufio.NewReader(conn)
	var response []byte

	role := replicaManager.Role
	replID := replicaManager.ReplID

	isMulti := false
	queue := [][]string{} 
	for {
		args, err :=resp.Parse(reader)
		if isMulti && strings.ToUpper(args[0]) != "EXEC" && strings.ToUpper(args[0]) != "DISCARD" {
			queue = append(queue, args)
			response = []byte("+QUEUED\r\n")
			conn.Write(response)
			continue
		}
		if err != nil {
			response = []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
			conn.Write(response)
			return
		}
		fmt.Printf("Received command: %v\n", args)
		switch strings.ToUpper(args[0]) {
		case "PING":
			response = handlePing()
			replicaManager.PropagateCommand(args)
		case "ECHO":
			response = handleEcho(args)
		case "SET":
			response = handleSet(st, args)
			replicaManager.PropagateCommand(args)
		case "GET":
			response = handleGet(st, args)
		case "RPUSH":
			response = handleRpush(st, args)
		case "LRANGE":
			response = handleLrange(st, args)
		case "LPUSH":
			response = handleLpush(st, args)
		case "LLEN":
			response = handleLlen(st, args)
		case "LPOP":	
			response = handleLpop(st, args)
		case "BLPOP":
			response = handleBLpop(st, args)
		case "TYPE":
			response = handleType(st, args)
		case "XADD":
			response = handleXAdd(st, args)
		case "XRANGE":
			response = handleXRange(st, args)
		case "XREAD":
			response = handleXRead(st, args)
		case "INCR":
			response = handleINCR(st, args)
		case "MULTI":
			isMulti = true
			queue = [][]string{}
			response = []byte("+OK\r\n")
		case "EXEC":
			if !isMulti {
				response = []byte("-ERR EXEC without MULTI\r\n")
			}else {
				response = handleEXEC(st, queue)
				isMulti = false
				queue = [][]string{}
			}
		case "DISCARD":
			if !isMulti {
				response = []byte("-ERR DISCARD without MULTI\r\n")
			} else {
				isMulti = false
				queue = [][]string{}
				response = []byte("+OK\r\n")
			}
		case "INFO":
			response = handleInfo(st, args, role, replID)
		case "REPLCONF":
			response = handleREPLCONF(st, args)
		case "PSYNC":
			response = handlePSYNC(st, args, replID)
			conn.Write(response)
			conn.Write(buildRDB())
			replicaManager.Add(conn)
			continue
		}
		PropagateCommands := []string{"SET", "PING"}
		for _, command := range PropagateCommands{
			if strings.ToUpper(args[0]) == command {
				replicaManager.PropagateCommand(args)
				break
			}
		}
		conn.Write(response)
	}
}
