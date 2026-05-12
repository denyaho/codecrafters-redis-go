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

func _sendPSYNC() []byte {
	return []byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n")
}

func HandleConnect_to_Master(conn net.Conn) {
	_, err := conn.Write(_sendPing())
	if err != nil {
		fmt.Printf("Failed to send PING to master: %v\n", err)
		return
	}
	isReplicationEstablished := false
	isFirstREPLCONFSent := false
	reader := bufio.NewReader(conn)
	for {
		args, err := resp.Parse(reader)
		if err != nil {
			fmt.Printf("Failed to read PING response from master: %v\n", err)
			return
		}
		switch strings.ToUpper(args[0]) {
		case "PONG":
			_, err := conn.Write(_sendREPLCONF(true))
			if err != nil {
				fmt.Printf("Failed to send REPLCONF to master: %v\n", err)
				return
			}
			isFirstREPLCONFSent = true
		case "OK":
			if isReplicationEstablished {
				_, err := conn.Write(_sendPSYNC())
				if err != nil {
					fmt.Printf("Failed to send PSYNC to master: %v\n", err)
					return
				}
			}
			if isFirstREPLCONFSent {
				_, err := conn.Write(_sendREPLCONF(false))
				if err != nil {
					fmt.Printf("Failed to send REPLCONF to master: %v\n", err)
					return
				}
				isReplicationEstablished = true
				isFirstREPLCONFSent = false
			}

		}
	}
}
