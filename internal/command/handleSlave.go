package handler

import (
	"fmt"
	"net"
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
	_, err = conn.Write(_sendREPLCONF(true))
	if err != nil {
		fmt.Printf("Failed to send REPLCONF listening-port to master: %v\n", err)
		return
	}
	_, err = conn.Write(_sendREPLCONF(false))
	if err != nil {
		fmt.Printf("Failed to send REPLCONF capa to master: %v\n", err)
		return
	}
}
