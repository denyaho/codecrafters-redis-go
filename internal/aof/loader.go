package aof

import (
	"fmt"
	"os"
	"sync"
)

type manifest struct {
	Filename string
	Sequence int64
	Type string
}

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFsync    string
	file *os.File
	mu sync.Mutex
	Manifest manifest
}

func NewAOF(dir string, appendOnly string, appendDirname string, appendFilename string, appendFsync string) *AOF {
	return &AOF{
		Dir: dir,
		AppendOnly: appendOnly,
		AppendDirname: appendDirname,
		AppendFilename: appendFilename,
		AppendFsync: appendFsync,
		Manifest: manifest{
			Filename: appendFilename + ".manifest",
			Sequence: 1,
			Type: "i",
		},
	}
}

func (a *AOF) CreateAOFDir() error {
	if a.AppendOnly == "yes" {
		if a.AppendDirname == "" {
			return os.MkdirAll(a.Dir + "/" + a.AppendFilename, 0755)
		}
		return os.MkdirAll(a.Dir + "/" + a.AppendDirname, 0755)	
	}
	return nil
}

func (a *AOF) GetAOFFilePath(filename string) string {
	if a.AppendDirname == "" {
		return a.Dir + "/" + filename
	}
	return a.Dir + "/" + a.AppendDirname + "/" + filename
}

func (a *AOF) Open() error {
	if a.AppendOnly != "yes" {
		return nil
	}
	f, err := os.OpenFile(a.GetAOFFilePath(a.AppendFilename+".1.incr.aof"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.file = f
	return nil
}

func _manifestContent(a *AOF) string {
	return fmt.Sprintf("file %s.1.incr.aof seq %d type %s", a.AppendFilename, a.Manifest.Sequence, a.Manifest.Type)
}

func (a *AOF) WriteManifest() error {
	if a.AppendOnly != "yes" {
		return nil
	}
	f, err := os.OpenFile(a.GetAOFFilePath(a.Manifest.Filename), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	content := _manifestContent(a)
	_, err = f.WriteString(content)
	return err
}

