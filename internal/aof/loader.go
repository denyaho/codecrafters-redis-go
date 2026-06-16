package aof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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


func (a *AOF) Sync() error {
	if a.AppendOnly != "yes" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Sync()
}

func (a *AOF) Write(args []byte) error {
	if a.AppendOnly != "yes" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.file.Write(args)
	return err
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
	filename, err := a.readManifestFile()
	if err != nil {
		return a.OpenFile(a.AppendFilename + ".1.incr.aof")
	}
	return a.OpenFile(filename)
}


func (a *AOF) OpenFile(filename string) error {
	f, err := os.OpenFile(a.GetAOFFilePath(filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.file = f
	return nil
}


func (a *AOF) readManifestFile() (string, error) {
	f, err := os.Open(a.Manifest.Filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var aofFile string
	for scanner.Scan() {
		line := scanner.Text()
		splitedLine := strings.Split(strings.TrimSpace(line), " ")
		filename := splitedLine[1]
		switch splitedLine[5] {
			case "i":
				aofFile = filename
		}
	}
	return aofFile, nil
}


func manifestContent(a *AOF) string {
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
	
	content := manifestContent(a)
	_, err = f.WriteString(content)
	return err
}

