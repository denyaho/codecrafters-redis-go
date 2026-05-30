package handler

import (
	"fmt"
	"strings"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"strconv"
	"time"
)

func handlePing() []byte {
	return resp.BuildSimpleString("PONG")
}

func handleEcho(args []string) []byte {
	return resp.BuildBulkStrings(args[1])
}

func handleSet(st *store.ExpireMap, args []string) []byte {
	fmt.Printf("handling set command with an value %v\n", args[2])
	if len(args) >= 4 && strings.ToUpper(args[3]) == "PX" {
				expireMs, _ := strconv.Atoi(args[4])
				expireAt := time.Duration(expireMs) * time.Millisecond
				st.Set(args[1], args[2], expireAt)
	}else if len(args) >= 4 && strings.ToUpper(args[3]) == "EX" {
		expireS, _ := strconv.Atoi(args[4])
		expireAt := time.Duration(expireS) * time.Second
		st.Set(args[1], args[2], expireAt)
	}else{
		st.Set(args[1], args[2], 0)
	}
	return resp.BuildSimpleString("OK")
}

func handleGet(st *store.ExpireMap, args []string) []byte {

	if val, ok := st.Get(args[1]); ok {
		return resp.BuildBulkStrings(val.(string))
	} else {
		return resp.BuildNullBulkString()
	}
}

func handleRpush(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.Rpush(args[1], args[2:]...)
	if err != nil {
		return resp.BuildError(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return resp.BuildInteger(list_length)
}

func handleLrange(st *store.ExpireMap, args []string) []byte {
	start := args[2]
	end := args[3]
	startIndex, _ := strconv.Atoi(start)
	endIndex, _ := strconv.Atoi(end)
	elem, err := st.Lrange(args[1], startIndex, endIndex)
	if err != nil {
		return resp.BuildError(fmt.Sprintf("%s\r\n", err.Error()))
	}
	if len(elem) == 0 {
		return resp.BuildArray([]string{})
	}
	return resp.BuildArray(elem)
}

func handleLpush(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.LPush(args[1], args[2:]...)
	if err != nil {
		return resp.BuildError(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return resp.BuildInteger(list_length)
}

func handleLlen(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.Llen(args[1])
	if err != nil {
		return resp.BuildError(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return resp.BuildInteger(list_length)
}

func handleLpop(st *store.ExpireMap, args []string) []byte {
	if len(args) == 2{
		val, ok, err := st.Lpop(args[1])
		if err != nil {
			return resp.BuildError(fmt.Sprintf("%s", err.Error()))
		}
		if ok {
			return resp.BuildBulkStrings(val)
		} else {
			return resp.BuildNullBulkString()
		}
	}else {
		popped_word := []string{}
		count, _ := strconv.Atoi(args[2])
		for i := 0; i < count; i++ {
			val, ok, err := st.Lpop(args[1])
			if err != nil {
				return resp.BuildError(fmt.Sprintf("%s\r\n", err.Error()))
			}
			if !ok {
				return resp.BuildNullBulkString()
			}
			popped_word = append(popped_word, val)
		}
		return resp.BuildArray(popped_word)
	}
}

func handleBLpop(st *store.ExpireMap, args []string) []byte {
	fmt.Println("handling BLPOP command")
	if len(args) < 3 {
		return resp.BuildError("ERR wrong number of arguments for 'BLPOP' command")
	}
	timeoutSec, _ := strconv.ParseFloat(args[2], 64)
	fmt.Printf("Parsed timeout: %f seconds\n", timeoutSec)
	timeout := time.Duration(timeoutSec * float64(time.Second))
	response, ok, istimeout := st.BLPop(args[1], timeout)
	fmt.Printf("BLPop result: response=%s, ok=%v, istimeout=%v\n", response, ok, istimeout)
	if istimeout {
		return resp.BuildNullArray()
	} else if !ok {
		return resp.BuildNullArray()
	}
	return resp.BuildArray([]string{args[1], response})
}
