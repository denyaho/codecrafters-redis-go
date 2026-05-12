package handler

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func _sendPing() []byte {
	return []byte("*1\r\n$4\r\nPING\r\n")
}

func _sendREPLCONF(isFirst bool) []byte {
	if isFirst {
		return []byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n")
	}
	return []byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")
}

func HandleConnect_to_Master(conn net.Conn) {
	_, err := conn.Write(_sendPing())
	if err != nil {
		fmt.Printf("Failed to send PING to master: %v\n", err)
		return
	}
	reader := bufio.NewReader(conn)
	for {
		args, err := resp.Parse(reader)
		if err != nil {
			fmt.Printf("Failed to read PING response from master: %v\n", err)
			return
		}
		switch strings.ToUpper(args[0]) {
		case "PONG":
			conn.Write(_sendREPLCONF(true))
			conn.Write(_sendREPLCONF(false))
		}
	}
}
