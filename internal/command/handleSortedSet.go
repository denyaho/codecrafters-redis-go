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
	member := args[3]
	score, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.BuildError("ERR value is not a valid float")
	}
	val, err := st.ZAdd(key, score, member)
	if err != nil {
		return resp.BuildError(err.Error())
	}
	return resp.BuildInteger(val)
}
