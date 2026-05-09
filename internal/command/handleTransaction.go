package handler

import (
	"strconv"
	
	"github.com/codecrafters-io/redis-starter-go/internal/store"
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