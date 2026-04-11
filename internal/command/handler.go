package handler

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)


func HandleConnection(conn net.Conn, st store.Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		args, err :=resp.Parse(reader)
		if err != nil {
			fmt.Println("failed to read from connection: ", err.Error())
			return
		}
		switch strings.ToUpper(args[0]) {
		case "ECHO":
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])))
		case "SET":
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
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if val, ok := st.Get(args[1]); ok {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
			} else {
				conn.Write([]byte("$-1\r\n"))
			}
		case "RPUSH":
			list_length := st.Rpush(args[1], args[2:]...)
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", list_length)))
		case "LRANGE":
			start := args[2]
			end := args[3]
			startIndex, _ := strconv.Atoi(start)
			endIndex, _ := strconv.Atoi(end)
			elem := st.Lrange(args[1], startIndex, endIndex)
			if len(elem) == 0 {
				conn.Write([]byte("*0\r\n"))
				continue
			}
			response := []byte(fmt.Sprintf("*%d\r\n", len(elem)))
			for i := 0; i < len(elem); i++ {
				word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(elem[i]), elem[i]))
				response = append(response, word...)
			}
			conn.Write(response)
		case "LPUSH":
			list_length := st.LPush(args[1], args[2:]...)
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", list_length)))
		case "LLEN":
			list_length := st.Llen(args[1])
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", list_length)))
		case "LPOP":
			if len(args) == 2{
				if val, ok := st.Lpop(args[1]); ok {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
				} else {
					conn.Write([]byte("$-1\r\n"))
				}
			}else {
				popped_word := []string{}
				count, _ := strconv.Atoi(args[2])
				for i := 0; i < count; i++ {
					if val, ok := st.Lpop(args[1]); ok {
						popped_word = append(popped_word, val)
					} else {
						conn.Write([]byte("$-1\r\n"))
						break
					}
				}
				if len(popped_word) == count {
					response := []byte(fmt.Sprintf("*%d\r\n", len(popped_word)))
					for i := 0; i < len(popped_word); i++ {
						word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(popped_word[i]), popped_word[i]))
						response = append(response, word...)
					}
					conn.Write(response)
				} else {
					conn.Write([]byte("$-1\r\n"))
				}
			}
		default:
				conn.Write([]byte("+PONG\r\n"))
		}
	}
}
