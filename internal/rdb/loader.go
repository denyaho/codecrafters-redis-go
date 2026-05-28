package rdb

import (
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type DBInfo struct {
	DBIndex int
	HashTableSize int
	ExpireTableSize int
}

type RDB struct {
	Dir string
	DBfilename string
	version int
	metadata map[string]string
	checksum []byte
	DBs []DBInfo
}

func NewRDB(dir, dbfilename string) *RDB {
	return &RDB{
		Dir: dir,
		DBfilename: dbfilename,
		metadata: make(map[string]string),
		version : 0,
		DBs: make([]DBInfo, 0),
	}
}


func (r *RDB) ReadFile(st *store.ExpireMap) error {
	data, err := os.ReadFile(r.Dir + "/" + r.DBfilename)
	if err != nil {
		return err
	}
	p := NewRDBParser(data)
	return p.Parse(r, st)
}

