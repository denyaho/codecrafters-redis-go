package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleSUBSCRIBE(args []string) []byte {
	response := resp.BuildArray([]string{"subscribe", args[1]})
	response = append(response, resp.BuildInteger(1)...)
	return response
}
