package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func handleType(st *store.ExpireMap, args []string) []byte {
	if len(args) < 2 {
		return []byte("-ERR wrong number of arguments for 'TYPE' command\r\n")
	}
	value, ok := st.Get(args[1])
	if !ok {
		return []byte("+none\r\n")
	}
	switch value.(type) {
	case string:
		return []byte("+string\r\n")
	case []string:
		return []byte("+list\r\n")
	default:
		return []byte("+none\r\n")
	}
}
