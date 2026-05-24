package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func handleCONFIG(args []string, rdbConfig *rdb.RDB) []byte {
	if args[1] == "GET" {
		switch args[2] {
		case "dir":
			return resp.BuildBulkStrings(rdbConfig.Dir)
		case "dbfilename":
			return resp.BuildBulkStrings(rdbConfig.DBfilename)
		}
	}
	return resp.BuildError("ERR Unsupported CONFIG subcommand")
}
