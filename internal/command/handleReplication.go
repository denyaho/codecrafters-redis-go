package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
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

func handleREPLCONF(st *store.ExpireMap, args []string) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'REPLCONF' command\r\n")
	}
	if args[1] == "listening-port" {
		return []byte("+OK\r\n")
	} else if args[1] == "capa" {
			return []byte("+OK\r\n")
	}
	return []byte("-ERR unknown REPLCONF option\r\n")
}

func handlePSYNC(st *store.ExpireMap, args []string) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'PSYNC' command\r\n")
	}
	replID := args[1]
	offset := args[2]
	return []byte(fmt.Sprintf("+FULLRESYNC %s %s\r\n", replID, offset))	
	
}
