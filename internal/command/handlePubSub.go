package handler

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/client"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleSUBSCRIBE(args []string, c *client.Client) []byte {
	c.Subscribe(args[1])
	return resp.BuildArrayForPUBSUB([]string{"subscribe", args[1]}, c.SubscriptionCount)
}


func handleSubscribedMode(c *client.Client, args []string) []byte {

	switch strings.ToUpper(args[0]) {
		case "SUBSCRIBE":
		case "UNSUBSCRIBE":
		case "PSUBSCRIBE":
		case "PUNSUBSCRIBE":
		case "PING":
		case "QUIT":
			
	}
	return resp.BuildError(fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context ", args[0]))
}