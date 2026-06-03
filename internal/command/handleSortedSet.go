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

	index, err := st.ZRank(args[1], args[2])
	if err != nil {
		return resp.BuildError(err.Error())
	}
	if index == -1 {
		return resp.BuildNullBulkString()
	}
	return resp.BuildInteger(index)
}

func handleZRANGE(st *store.ExpireMap, args []string) []byte {
	if len(args) != 4 {
		return resp.BuildError("ERR wrong number of arguments for 'ZRANGE' command")
	}
	key := args[1]
	start, err1 := strconv.Atoi(args[2])
	end, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.BuildError("ERR value is not an integer")
	}
	members, err := st.ZRange(key, start, end)
	if err != nil {
		return resp.BuildError(err.Error())
	}
	return resp.BuildArray(members)

}

func handleZCARD(st *store.ExpireMap, args []string) []byte {
	if len(args) != 2 {
		return resp.BuildError("ERR wrong number of arguments for 'ZCARD' command")
	}
	key := args[1]
	card, err := st.ZCard(key)
	if err != nil {
		return resp.BuildError(err.Error())
	}
	return resp.BuildInteger(card)
}

func handleZSCORE(st *store.ExpireMap, args []string) []byte {
	if len(args) != 3 {
		return resp.BuildError("ERR wrong number of arguments for 'ZSCORE' command")
	}
	key := args[1]
	member := args[2]
	score, err := st.ZScore(key, member)
	if err != nil {
		return resp.BuildError(err.Error())
	}
	if score == -1 {
		return resp.BuildNullBulkString()
	}
	return resp.BuildBulkStrings(strconv.FormatFloat(score, 'f', -1, 64))
}