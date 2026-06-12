package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
	"strings"
	"crypto/sha256"
	"encoding/hex"
)

func handleACL(st *store.ExpireMap, args []string, ps *pubsub.Manager, c *pubsub.Client) []byte {

	switch args[1] {
	case "WHOAMI":
		return handleACLWhoami(st, args, ps, c)
	case "GETUSER":
		return handleACLGetUser(st, args, ps, c)
	case "SETUSER":
		return handleACLSetUser(st, args, ps, c)
	}
	return resp.BuildError("ERR Unsupported ACL subcommand")
}

func handleACLWhoami(st *store.ExpireMap, args []string, ps *pubsub.Manager, c *pubsub.Client) []byte {
	return resp.BuildBulkStrings(ps.Users[c.Username].Username)
}
func handleACLGetUser(st *store.ExpireMap, args []string, ps *pubsub.Manager, c *pubsub.Client) []byte {
	user := args[2]
	if ps.GetUser(user) == nil {
		return resp.BuildNullArray()
	}
	response := []byte("*4\r\n")
	response = append(response, resp.BuildBulkStrings("flags")...)
	response = append(response, resp.BuildArray(ps.Users[user].Flags)...)
	response = append(response, resp.BuildBulkStrings("passwords")...)
	response = append(response, resp.BuildArray(ps.Users[user].Passwords)...)
	return response
}

func handleACLSetUser(st *store.ExpireMap, args []string, ps *pubsub.Manager, c *pubsub.Client) []byte {
	username := args[2]
	for i := 3; i < len(args); i++ {
		if strings.HasPrefix(args[i], ">"){
			password := strings.TrimPrefix(args[i], ">")
			if user, exists := ps.Users[username];
			exists {
				hash := sha256.Sum256([]byte(password))
				user.Passwords = append(user.Passwords, hex.EncodeToString(hash[:]))
				flags := []string{}
				for _, flag := range user.Flags {
					if flag != "nopass" {
						flags = append(flags, flag)
					}
				}
				user.Flags = flags
			}
		}
	}
	return resp.BuildSimpleString("OK")
}

