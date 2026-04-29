package handler

import (
	"fmt"
	"strings"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"strconv"
	"time"
)

func handlePing() []byte {
	return []byte("+PONG\r\n")
}

func handleEcho(args []string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1]))
}

func handleSet(st *store.ExpireMap, args []string) []byte {
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
	return []byte("+OK\r\n")
}

func handleGet(st *store.ExpireMap, args []string) []byte {
	if val, ok := st.Get(args[1]); ok {
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val.(string)), val.(string)))
	} else {
		return []byte("$-1\r\n")
	}
}

func handleRpush(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.Rpush(args[1], args[2:]...)
	if err != nil {
		return []byte(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return []byte(fmt.Sprintf(":%d\r\n", list_length))
}

func handleLrange(st *store.ExpireMap, args []string) []byte {
	start := args[2]
	end := args[3]
	startIndex, _ := strconv.Atoi(start)
	endIndex, _ := strconv.Atoi(end)
	elem, err := st.Lrange(args[1], startIndex, endIndex)
	if err != nil {
		return []byte(fmt.Sprintf("%s\r\n", err.Error()))
	}
	if len(elem) == 0 {
		return []byte("*0\r\n")
	}
	response := []byte(fmt.Sprintf("*%d\r\n", len(elem)))
	for i := 0; i < len(elem); i++ {
		word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(elem[i]), elem[i]))
		response = append(response, word...)
	}
	return []byte(response)
}

func handleLpush(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.LPush(args[1], args[2:]...)
	if err != nil {
		return []byte(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return []byte(fmt.Sprintf(":%d\r\n", list_length))
}

func handleLlen(st *store.ExpireMap, args []string) []byte {
	list_length, err := st.Llen(args[1])
	if err != nil {
		return []byte(fmt.Sprintf("%s\r\n", err.Error()))
	}
	return []byte(fmt.Sprintf(":%d\r\n", list_length))
}

func handleLpop(st *store.ExpireMap, args []string) []byte {
	if len(args) == 2{
		val, ok, err := st.Lpop(args[1])
		if err != nil {
			return []byte(fmt.Sprintf("%s\r\n", err.Error()))
		}
		if ok {
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
		} else {
			return []byte("$-1\r\n")
		}
	}else {
		popped_word := []string{}
		count, _ := strconv.Atoi(args[2])
		for i := 0; i < count; i++ {
			val, ok, err := st.Lpop(args[1])
			if err != nil {
				return []byte(fmt.Sprintf("%s\r\n", err.Error()))
			}
			if !ok {
				return []byte("$-1\r\n")
			}
			popped_word = append(popped_word, val)
		}
		response := []byte(fmt.Sprintf("*%d\r\n", len(popped_word)))
		for i := 0; i < len(popped_word); i++ {
			word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(popped_word[i]), popped_word[i]))
			response = append(response,word...)
		}
		return response
	}
}

func handleBLpop(st *store.ExpireMap, args []string) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'BLPOP' command\r\n")
	}
	timeoutSec, _ := strconv.Atoi(args[2])
	timeout := time.Duration(timeoutSec) * time.Second
	response, ok, istimeout := st.BLPop(args[1], timeout)
	fmt.Printf("BLPop result: response=%s, ok=%v, istimeout=%v\n", response, ok, istimeout)
	if istimeout {
		return []byte("*-1\r\n")
	} else if !ok {
		return []byte("$-1\r\n")
	}
	return []byte(fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", 2, len(args[1]), args[1], len(response), response))
}
