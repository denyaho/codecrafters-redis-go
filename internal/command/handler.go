package handler

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/pubsub"
)


func HandleConnection(c *pubsub.Client, st *store.ExpireMap, replicaManager *replication.ReplicaManager, rdbConfig *rdb.RDB, ps *pubsub.Manager) {
	defer c.Connection.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	reader := bufio.NewReader(c.Connection)
	var response []byte

	role := replicaManager.Role
	replID := replicaManager.ReplID

	isMulti := false
	queue := [][]string{} 
	for {
		args, err :=resp.Parse(reader)
		if c.IsSubscribed {
			response = handleSubscribedMode(c, args, ps)
			c.Connection.Write(response)
			continue
		}
		command := strings.ToUpper(args[0])

		if isMulti && command == "WATCH" {
			response = []byte("-ERR WATCH inside MULTI is not allowed\r\n")
			c.Connection.Write(response)
			continue
		}
		if isMulti && command != "EXEC" && command != "DISCARD" {
			queue = append(queue, args)
			response = []byte("+QUEUED\r\n")
			c.Connection.Write(response)
			continue
		}
		if err != nil {
			response = []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
			c.Connection.Write(response)
			return
		}
		fmt.Printf("Received command: %v\n", args)
		switch command {
		case "PING":
			response = handlePing()
		case "ECHO":
			response = handleEcho(args)
		case "SET":
			response = handleSet(st, args)
		case "GET":
			response = handleGet(st, args)
		case "RPUSH":
			response = handleRpush(st, args)
		case "LRANGE":
			response = handleLrange(st, args)
		case "LPUSH":
			response = handleLpush(st, args)
		case "LLEN":
			response = handleLlen(st, args)
		case "LPOP":	
			response = handleLpop(st, args)
		case "BLPOP":
			response = handleBLpop(st, args)
		case "TYPE":
			response = handleType(st, args)
		case "XADD":
			response = handleXAdd(st, args)
		case "XRANGE":
			response = handleXRange(st, args)
		case "XREAD":
			response = handleXRead(st, args)
		case "INCR":
			response = handleINCR(st, args)
		case "MULTI":
			isMulti = true
			queue = [][]string{}
			response = []byte("+OK\r\n")
		case "EXEC":
			if !isMulti {
				response = []byte("-ERR EXEC without MULTI\r\n")
			}else {
				response = handleEXEC(st, queue)
				isMulti = false
				queue = [][]string{}
			}
		case "DISCARD":
			if !isMulti {
				response = []byte("-ERR DISCARD without MULTI\r\n")
			} else {
				isMulti = false
				queue = [][]string{}
				response = []byte("+OK\r\n")
			}
		case "WATCH":
			response = handleWATCH(st, args)
			
		case "INFO":
			response = handleInfo(st, args, role, replID)
		case "REPLCONF":
			response = handleREPLCONF(st, args, replicaManager)
		case "PSYNC":
			response = handlePSYNC(st, args, replID)
			c.Connection.Write(response)
			c.Connection.Write(resp.BuildRDB())
			replicaManager.Add(c.Connection)
			continue
		case "WAIT":
			response = handleWAIT(args, replicaManager)
		case "CONFIG":
			response = handleCONFIG(args, rdbConfig)
		case "KEYS":
			response = handleKEY(args, rdbConfig,st)
		case "SUBSCRIBE":
			response = handleSUBSCRIBE(args[1], c, ps)
		case "PUBLISH":
			response = handlePUBLISH(args[1], args[2], ps)
		case "ZADD":
			response = handleZADD(st, args)
		case "ZRANK":
			response = handleZRANK(st, args)
		case "ZRANGE":
			response = handleZRANGE(st, args)
		case "ZCARD":
			response = handleZCARD(st, args)
		case "ZSCORE":
			response = handleZSCORE(st, args)
		case "ZREM":
			response = handleZREM(st, args)
		case "GEOADD":
			response = handleGEOADD(st, args)
		case "GEOPOS":
			response = handleGEOPOS(st, args)	
		case "GEODIST":
			response = handleGEODIST(st, args)
		case "GEOSEARCH":
			response = handleGEOSEARCH(st, args)
		case "ACL":
			response = handleACL(st, args, ps, c)
		case "AUTH":
			response = handleAUTH(st, args, ps, c)
		}
		PropagateCommands := []string{"SET"}
		for _, command := range PropagateCommands{
			if strings.ToUpper(args[0]) == command {
				replicaManager.PropagateCommand(args)
			}
		}
		fmt.Printf("Sending response: %s\n", string(response))
		c.Connection.Write(response)
	}
}
