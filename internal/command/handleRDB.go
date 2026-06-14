package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/aof"
)


func handleCONFIG(args []string, rdbConfig *rdb.RDB, aofConfig *aof.AOF) []byte {
	if args[1] == "GET" {
		switch args[2] {
		case "dir":
			return resp.BuildArray([]string{"dir", rdbConfig.Dir})
		case "dbfilename":
			return resp.BuildArray([]string{"dbfilename", rdbConfig.DBfilename})
		case "appendonly":
			return resp.BuildArray([]string{"appendonly", aofConfig.AppendOnly})
		case "appenddirname":
			return resp.BuildArray([]string{"appenddirname", aofConfig.AppendDirname})
		case "appendfilename":
			return resp.BuildArray([]string{"appendfilename", aofConfig.AppendFilename})
		case "appendfsync":
			return resp.BuildArray([]string{"appendfsync", aofConfig.AppendFsync})
		}
	}
	return resp.BuildError("ERR Unsupported CONFIG subcommand")
}


func handleKEY(args []string, rdbConfig *rdb.RDB, st *store.ExpireMap) []byte {
	if len(args) != 2 {
		return resp.BuildError("ERR wrong number of arguments for 'KEY' command")
	}
	keys := st.Keys(args[1])
	return resp.BuildArray(keys)
}
