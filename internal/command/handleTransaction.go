package handler

import (
	"strconv"
	"strings"	
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
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
	return resp.BuildInteger(intValue)
}

func _checkWatchedKeys(st *store.ExpireMap, c *pubsub.Client) bool {
	for key, version := range c.Watchedkeys {
		fmt.Printf("Checking key: %s, version: %d, current version: %d\n", key, version, st.GetVersion(key))

		if st.GetVersion(key) != version {
			return false
		}
	}
	return true
}

func handleEXEC(st *store.ExpireMap, queue [][]string, c *pubsub.Client) []byte {
	var response []byte
	if len(queue) == 0 {
		return []byte("*0\r\n")
	}
	if !_checkWatchedKeys(st, c) {
		return []byte("*-1\r\n")
	}
	responses := []byte(fmt.Sprintf("*%d\r\n", len(queue)))
	for _, args := range queue {
		command := strings.ToUpper(args[0])
		
		switch command {
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

func handleWATCH(st *store.ExpireMap, args []string) []byte {
	return []byte("+OK\r\n")
}