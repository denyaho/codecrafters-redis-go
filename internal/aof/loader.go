package aof

import (
	"os"
)

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFsync    string
}

func NewAOF(dir string, appendOnly string, appendDirname string, appendFilename string, appendFsync string) *AOF {
	return &AOF{
		Dir: dir,
		AppendOnly: appendOnly,
		AppendDirname: appendDirname,
		AppendFilename: appendFilename,
		AppendFsync: appendFsync,
	}
}

func (a *AOF) CreateDir() error {
	if a.AppendOnly == "yes" {
		return os.Mkdir(a.Dir + "/" + a.AppendDirname, 0755)	
	}
	return nil
}