package main

import (
	"fmt"
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func initRDB(config *Config, st *store.ExpireMap) (*rdb.RDB, error) {
	rdb := rdb.NewRDB(config.CurrentDir, config.DBFilename)

	if err := rdb.ReadFile(st); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("RDB file not found, starting with an empty database.")
		} else {
			fmt.Printf("Failed to read RDB file: %v\n", err)
			os.Exit(1)
		}
	}
	return rdb, nil
}
