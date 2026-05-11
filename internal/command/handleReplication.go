package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
	"net"
)

func handleInfo(st *store.ExpireMap, args []string, role, replID string) []byte {
	fields := [][2]string{
		{"role", role},
		{"master_replid", replID},
		{"connected_slaves", "0"},
		{"master_repl_offset", "0"},
		{"second_repl_offset", "-1"},
		{"repl_backlog_active", "0"},
		{"repl_backlog_size", "1048576"},
		{"repl_backlog_first_byte_offset", "0"},
		{"repl_backlog_histlen", "0"},
	}

	body := ""
	for _, kv := range fields {
		body += fmt.Sprintf("%s:%s\r\n", kv[0], kv[1])
	}
	response := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(body), body))
	return response
}

func _sendPing() []byte {
	return []byte("*1\r\n$4\r\nPING\r\n")
}

func HandleConnect_to_Master(conn net.Conn) {
	_, err := conn.Write(_sendPing())
	if err != nil {
		fmt.Printf("Failed to send PING to master: %v\n", err)
		return
	}
}