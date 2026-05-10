package handler

import (
	"strconv"
	"strings"	
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
)

func handleINCR(st *store.ExpireMap, args []string) []byte {
	if len(args) != 2 {
		return []byte("-ERR wrong number of arguments for 'INCR' command\r\n")
	}
	key := args[1]
	val, ok := st.Get(key)
	if !ok {
		st.Set(key, "1", 0)
		return []byte(":1\r\n")
	}
	intValue,err := strconv.Atoi(val.(string))
	if err != nil {
		return []byte("-ERR value is not an integer or out of range\r\n")
	}
	intValue++
	st.Set(key, strconv.Itoa(intValue), 0)
	return []byte(":" + strconv.Itoa(intValue) + "\r\n")
}

func handleEXEC(st *store.ExpireMap, queue [][]string) []byte {
	var response []byte
	if len(queue) == 0 {
		return []byte("*0\r\n")
	}
	responses := []byte(fmt.Sprintf("*%d\r\n", len(queue)))
	for i, args := range queue {
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
		case "XREAD":
			response = handleXRead(st, args)
		case "INCR":
			response = handleINCR(st, args)
		}
		responses = append(responses, response...)
	}
	return responses
}