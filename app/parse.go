package main

import (
	"os"
	"strings"
)

type Config struct {
	Port       string
	MasterPort string
	MasterAddr string
	Role       string
	CurrentDir string
	DBFilename string
	AppendFilename string
	AppendFsync string
	AppendOnly string
	AppendDirname string
}

func parseArgs() Config {

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config := Config{
		Port:       "6379",
		MasterPort: "",
		MasterAddr: "",
		Role:       "master",
		CurrentDir: dir,
		DBFilename: "",
		AppendFilename: "",
		AppendFsync: "",
		AppendOnly: "no",
		AppendDirname: "",
	}

	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			config.Port = args[i+1]
		case "--replicaof":
			config.Role = "slave"
			parts := strings.Split(args[i+1], " ")
			config.MasterAddr = parts[0]
			config.MasterPort = parts[1]
		case "--dir":
			config.CurrentDir = args[i+1]
		case "--dbfilename":
			config.DBFilename = args[i+1]
		case "--appendonly":
			config.AppendOnly = args[i+1]
		case "--appenddirname":
			config.AppendDirname = args[i+1]
		case "--appendfilename":
			config.AppendFilename = args[i+1]
		case "--appendfsync":
			config.AppendFsync = args[i+1]
		}
	}
	return config
}
