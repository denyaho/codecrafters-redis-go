package handler

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

func _sendPing() []byte {
	return resp.BuildArray([]string{"PING"})
}

func _sendREPLCONF(isFirst bool) []byte {
	if isFirst {
		return resp.BuildArray([]string{"REPLCONF", "listening-port", "6380"})
	}
	return resp.BuildArray([]string{"REPLCONF", "capa", "psync2"})
}

func _sendPSYNC() []byte {
	return resp.BuildArray([]string{"PSYNC", "?", "-1"})
}

func _respondACK(offset int64) []byte {
	return replication.BuildCommand([]string{"REPLCONF", "ACK", fmt.Sprintf("%d", offset)})
}

func HandleConnect_to_Master(conn net.Conn, st *store.ExpireMap, replicaManager *replication.ReplicaManager) {
	_, err := conn.Write(_sendPing())
	if err != nil {
		fmt.Printf("Failed to send PING to master: %v\n", err)
		return
	}
	isReplicationEstablished := false
	isFirstREPLCONFSent := false
	reader := bufio.NewReader(conn)
	for {
		args, err := resp.Parse(reader)
		if err != nil {
			fmt.Printf("Failed to read PING response from master: %v\n", err)
			return
		}
		if strings.HasPrefix(string(args[0]), "FULLRESYNC") {
			replicaManager.SetPsynced(true)
			continue
		}
		switch strings.ToUpper(args[0]) {
		case "PONG":
			_, err := conn.Write(_sendREPLCONF(true))
			if err != nil {
				fmt.Printf("Failed to send REPLCONF to master: %v\n", err)
				return
			}
			isFirstREPLCONFSent = true
		case "OK":
			if isReplicationEstablished {
				_, err := conn.Write(_sendPSYNC())
				if err != nil {
					fmt.Printf("Failed to send PSYNC to master: %v\n", err)
					return
				}
			}
			if isFirstREPLCONFSent {
				_, err := conn.Write(_sendREPLCONF(false))
				if err != nil {
					fmt.Printf("Failed to send REPLCONF to master: %v\n", err)
					return
				}
				isReplicationEstablished = true
				isFirstREPLCONFSent = false
			}
		case "SET":
			_ = handleSet(st, args)
			replicaManager.AddOffset(args)
		case "PING":
			_ = handlePing()
			replicaManager.AddOffset(args)
		case "REPLCONF":
			if args[1] == "GETACK" && args[2] == "*" {
				_, err := conn.Write(_respondACK(replicaManager.GetOffset()))
				if err != nil {
					fmt.Printf("Failed to send ACK to master: %v\n", err)
					return
				}
				replicaManager.AddOffset(args)
			}		
		}
	}
}
