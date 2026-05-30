package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/client"
)

func handleSUBSCRIBE(args []string, c *client.Client) []byte {
	c.Subscribe(args[1])
	return resp.BuildArrayForPUBSUB([]string{"subscribe", args[1]}, c.SubscriptionCount)
}
