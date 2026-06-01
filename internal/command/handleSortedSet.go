package handler

import (
	"strconv"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"

)

func handleZADD(st *store.ExpireMap, args []string) []byte {
	if len(args) != 4 {
		return resp.BuildError("ERR wrong number of arguments for 'ZADD' command")
	}

	key := args[1]
	score, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.BuildError("ERR value is not a valid float")
	}
	member := args[3]
	val, err := st.ZAdd(key, score, member)
	if err != nil {
		return resp.BuildError(err.Error())
	}
	return resp.BuildInteger(val)
}

func handleZRANK(st *store.ExpireMap, args []string) []byte {
	if len(args) != 3 {
		return resp.BuildError("ERR wrong number of arguments for 'ZRANK' command")
	}
	key := args[1]
	member := args[2]
	item, exist := st.Get(key)
	if !exist {
		return resp.BuildError("ERR no such key")
	}
	zset, ok := item.([]store.ZSetEntry)
	if !ok {
		return resp.BuildError("ERR wrong type of value for 'ZRANK' command")
	}
	for i, entry := range zset {
		if entry.Member == member {
			return resp.BuildInteger(i)
		}
	}
	return resp.BuildError("ERR member not found")
}