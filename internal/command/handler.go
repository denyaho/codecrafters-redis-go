package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)


func HandleConnection(conn net.Conn, st *store.ExpireMap) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var response []byte
	for {
		args, err :=resp.Parse(reader)
		if err != nil {
			response = []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
			conn.Write(response)
			return
		}
		fmt.Printf("Received command: %v\n", args)
		switch strings.ToUpper(args[0]) {
		case "PING":
			response = handlePing()
		case "ECHO":
			response = handleEcho(args)
		case "SET":
			response = handleSet(st, args)			
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
		}
		conn.Write(response)
	}
}
