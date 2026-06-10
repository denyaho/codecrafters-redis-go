package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
)

func handleACL(st *store.ExpireMap, args []string, c *pubsub.Client) []byte {

	switch args[1] {
	case "WHOAMI":
		return handleACLWhoami(st, args, c)
	case "GETUSER":
		return handleACLGetUser(st, args, c)
	}
	return resp.BuildError("ERR Unsupported ACL subcommand")
}

func handleACLWhoami(st *store.ExpireMap, args []string, c *pubsub.Client) []byte {
	return resp.BuildBulkStrings(c.Userinfo.Username)
}

func handleACLGetUser(st *store.ExpireMap, args []string, c *pubsub.Client) []byte {
	response := []byte("*3\r\n")
	response = append(response, resp.BuildBulkStrings("flags")...)
	response = append(response, resp.BuildArray([]string{"nopass"})...)
	response = append(response, resp.BuildBulkStrings("passwords")...)
	response = append(response, resp.BuildArray([]string{})...)
	return response
}


