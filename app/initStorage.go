package main

import (
	"github.com/codecrafters-io/redis-starter-go/internal/aof"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"os"
	"fmt"
	"time"
)

func initStorage(config *Config) (*store.ExpireMap, *rdb.RDB, *aof.AOF) {

	st := store.NewExpireMap()

	rdb, err := initRDB(config, st)
	if err != nil {
		fmt.Printf("Failed to initialize RDB: %v\n", err)
		os.Exit(1)
	}

	if err := rdb.ReadFile(st); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("RDB file not found, starting with empty database")
		} else {
			fmt.Printf("Failed to load RDB file: %v\n", err)
			os.Exit(1)
		}
	}
	aof, err := initAOF(config, st)
	if err != nil {
		fmt.Printf("Failed to initialize AOF: %v\n", err)
		os.Exit(1)
	}


	if err := aof.CreateAOFDir(); err != nil {
		fmt.Printf("Failed to create AOF directory: %v\n", err)
		os.Exit(1)
	}

	if err := aof.OpenFile(aof.AppendFilename); err != nil {
		fmt.Printf("Failed to open AOF: %v\n", err)
		os.Exit(1)
	}

	if aof.AppendFsync == "everysec" {
		go func() {
			timer := time.NewTicker(1 * time.Second)
			for {
				select {
					case <-timer.C:
						if err := aof.Sync(); err != nil {
							fmt.Printf("Failed to sync AOF file: %v\n", err)
						}
					}
				}
			}()
	}
	if err := aof.WriteManifest(); err != nil {
		fmt.Printf("Failed to write AOF manifest: %v\n", err)
		os.Exit(1)
	}
	aof.Manifest.Sequence++

	return st, rdb, aof
}
