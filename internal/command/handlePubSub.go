package handler

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleSUBSCRIBE(args []string, c *pubsub.Client) []byte {
	c.Subscribe(args[1])
	return resp.BuildArrayForPUBSUB([]string{"subscribe", args[1]}, c.SubscriptionCount)
}

func handleUNSUBSCRIBE(args []string, c *pubsub.Client) []byte {
	c.Unsubscribe(args[1])
	return resp.BuildArrayForPUBSUB([]string{"unsubscribe", args[1]}, c.SubscriptionCount)
}




func handleSubscribedMode(c *pubsub.Client, args []string) []byte {
	switch strings.ToUpper(args[0]) {
		case "SUBSCRIBE":
			return handleSUBSCRIBE(args, c)
		case "UNSUBSCRIBE":
			return handleUNSUBSCRIBE(args, c)
		case "PSUBSCRIBE":
		case "PUNSUBSCRIBE":
		case "PING":
			return resp.BuildArray([]string{"pong", ""})
		case "QUIT":

	}
	return resp.BuildError(fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context ", args[0]))
}


func handlePUBLISH(channel, msg string, ps *pubsub.Manager) []byte {
	return ps.Publish(channel, msg)
}