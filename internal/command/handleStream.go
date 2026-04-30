package handler

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type StreamEntry struct {
	ID string
	value map[string]string
}

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

func handleXAdd(st *store.ExpireMap, args []string) []byte {
	if len(args) < 4 || len(args[3:])%2 != 0 {
		return []byte("-ERR wrong number of arguments for 'XADD' command\r\n")
	}

	key := args[1]
	entryID := args[2]
	pairs := make(map[string]string)
	for i := 3; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]
		pairs[field] = value
	}
	entry := StreamEntry{
		ID: entryID,
		value: pairs,
	}
	st.Set(key, entry, 0)
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n",len(entryID), entryID))
}
