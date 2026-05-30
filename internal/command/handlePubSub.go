package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleSUBSCRIBE(args []string) []byte {
	return resp.BuildArrayForPUBSUB([]string{"subscribe", args[1]}, 1)
}
