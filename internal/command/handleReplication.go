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
	for i := 1; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return []byte("-ERR wrong number of arguments for 'REPLCONF' command\r\n")
		}
		if args[i] == "listening-port" {
			return []byte("+OK\r\n")
		}
		if args[i] == "capa" {
			return []byte("+OK\r\n")
		}
	}
	return []byte("-ERR unknown REPLCONF option\r\n")
}

var emtpyRDB = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

func buildRDB() []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(emtpyRDB), emtpyRDB))
}

func handlePSYNC(st *store.ExpireMap, args []string, replID string) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'PSYNC' command\r\n")
	}
	offset := args[2]
	if offset == "-1" {
		offset = "0"
	}
	if args[1] == "?" {
		return []byte(fmt.Sprintf("+FULLRESYNC %s %s\r\n", replID, offset))
	}
	return []byte(fmt.Sprintf("+FULLRESYNC %s %s\r\n", replID, offset))	
}
