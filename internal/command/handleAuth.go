package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleACLWhoami(st *store.ExpireMap, args []string) []byte {
	return resp.BuildBulkStrings("default")
}
