package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"fmt"
)

func handleSUBSCRIBE(args []string) []byte {
	response := resp.BuildArray([]string{"subscribe", args[1], fmt.Sprintf(":%d", 1)})
	return response
}
