package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleACL(st *store.ExpireMap, args []string) []byte {

	switch args[1] {
	case "WHOAMI":
		return handleACLWhoami(st, args)
	}
	return resp.BuildError("ERR Unsupported ACL subcommand")
}

func handleACLWhoami(st *store.ExpireMap, args []string) []byte {
	return resp.BuildBulkStrings("default")
}
