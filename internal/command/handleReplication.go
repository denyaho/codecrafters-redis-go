package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
)

func handleInfo(st *store.ExpireMap, args []string, role, replID string) []byte {
	lines := make(map[string]string)
	lines["role"] = role
	lines["replid"] = replID
	lines["connected_slaves"] = "0"
	lines["connected_clients"] = "0"
	lines["used_memory"] = "0"
	lines["used_memory_human"] = "0B"
	lines["used_memory_peak"] = "0"
	lines["used_memory_peak_human"] = "0B"

	var response []byte
	for key, value := range lines{
		response = append(response, []byte(fmt.Sprintf("$%d\r\n%s:%s\r\n", len(key)+len(value)+1, key, value))...)
	}
	return response
}
