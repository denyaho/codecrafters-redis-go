package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"fmt"
)


func handleCONFIG(args []string, rdbConfig *rdb.RDB) []byte {
	if args[1] == "GET" {
		switch args[2] {
		case "dir":
			return resp.BuildArray([]string{"dir", rdbConfig.Dir})
		case "dbfilename":
			return resp.BuildArray([]string{"dbfilename", rdbConfig.DBfilename})
		}
	}
	return resp.BuildError("ERR Unsupported CONFIG subcommand")
}


func handleKEY(args []string, rdbConfig *rdb.RDB, st *store.ExpireMap) []byte {
	if len(args) != 2 {
		return resp.BuildError("ERR wrong number of arguments for 'KEY' command")
	}
	fmt.Printf("Loading RDB file: %s/%s\n", rdbConfig.Dir, rdbConfig.DBfilename)
	rdbConfig.ReadFile(st)
	fmt.Printf("RDB file loaded successfully\n")
	keys := st.Keys(args[1])
	return resp.BuildArray(keys)
}
