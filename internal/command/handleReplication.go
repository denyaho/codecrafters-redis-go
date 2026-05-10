package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
)

func handleInfo(st *store.ExpireMap, args []string, role, replID string) []byte {
	lines := make(map[string]string)
	lines["role"] = "master"
	lines["master_replid"] = replID
	lines["connected_slaves"] = "0"
	lines["master_repl_offset"] = "-1"
	lines["second_repl_offset"] = "-1"
	lines["repl_backlog_active"] = "0"
	lines["repl_backlog_size"] = "1048576"
	lines["repl_backlog_first_byte_offset"] = "0"
	lines["repl_backlog_histlen"] = "0"

	var response []byte
	for key, value := range lines{
		response = append(response, []byte(fmt.Sprintf("$%d\r\n%s:%s\r\n", len(key)+len(value)+1, key, value))...)
	}
	return response
}
