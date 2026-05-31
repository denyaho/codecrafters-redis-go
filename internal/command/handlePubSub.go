package handler

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleSUBSCRIBE(channel string, c *pubsub.Client, ps *pubsub.Manager) []byte {
	c.Subscribe(channel)
	ps.Subscribe(c, channel)
	return resp.BuildArrayForPUBSUB([]string{"subscribe", channel}, c.SubscriptionCount)
}

func handleUNSUBSCRIBE(channel string, c *pubsub.Client, ps *pubsub.Manager) []byte {
	c.Unsubscribe(channel)
	ps.Unsubscribe(c, channel)
	return resp.BuildArrayForPUBSUB([]string{"unsubscribe", channel}, c.SubscriptionCount)
}




func handleSubscribedMode(c *pubsub.Client, args []string, ps *pubsub.Manager) []byte {
	command := args[0]
	switch strings.ToUpper(command) {
		case "SUBSCRIBE":
			channel := args[1]
			return handleSUBSCRIBE(channel, c, ps)
		case "UNSUBSCRIBE":
			channel := args[1]
			return handleUNSUBSCRIBE(channel, c, ps)
		case "PSUBSCRIBE":
		case "PUNSUBSCRIBE":
		case "PING":
			return resp.BuildArrayForPing()
		case "QUIT":

	}
	return resp.BuildError(fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context ", args[0]))
}


func handlePUBLISH(channel, msg string, ps *pubsub.Manager) []byte {
	return ps.Publish(channel, msg)
}