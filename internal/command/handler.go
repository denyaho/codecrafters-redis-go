package handler

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)


func HandleConnection(conn net.Conn, st store.Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	elem := []string{}
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
			for i:=2; i < len(args); i++ {
				elem = append(elem, args[i])
			}
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", len(elem))))
		default:
				conn.Write([]byte("+PONG\r\n"))
		}
	}
}
