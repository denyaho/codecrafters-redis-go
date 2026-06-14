package aof

import (
	"os"
	"sync"
)

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFsync    string
	file *os.File
	mu sync.Mutex
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

func (a *AOF) CreateAOFDir() error {
	if a.AppendOnly == "yes" {
		return os.MkdirAll(a.Dir + "/" + a.AppendDirname, 0755)	
	}
	return nil
}

func (a *AOF) GetAOFFilePath() string {
	return a.Dir + "/" + a.AppendDirname + "/" + a.AppendFilename
}

func (a *AOF) Open() error {
	if a.AppendOnly != "yes" {
		return nil
	}
	f, err := os.OpenFile(a.GetAOFFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.file = f
	return nil
}
