package aof

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Manifest struct {
	AOFFilename string
	Sequence int64
	Type string
}

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFsync    string
	ManifestFilename string
	file *os.File
	mu sync.Mutex
	Manifest Manifest
}

func NewAOF(dir string, appendOnly string, appendDirname string, appendFilename string, appendFsync string) *AOF {
	return &AOF{
		Dir: dir,
		AppendOnly: appendOnly,
		AppendDirname: appendDirname,
		AppendFilename: appendFilename,
		AppendFsync: appendFsync,
		ManifestFilename: appendFilename + ".manifest",
		Manifest: Manifest{
			AOFFilename: appendFilename + ".1.incr.aof",
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

func (a *AOF) Exists(name string) (bool, error) {
	f, err := os.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	defer f.Close()
	return true, nil
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

// func (a *AOF) Open(filePath string) error {
// 	if a.AppendOnly != "yes" {
// 		return nil
// 	}
// 	aofFiles, err := a.readManifestFile(filePath)
// 	if err != nil {
// 		return a.OpenFile(a.AppendFilename + ".1.incr.aof")
// 	}
// 	return a.OpenFile(filename)
// }


func (a *AOF) OpenFile(filename string) error {
	f, err := os.OpenFile(a.GetAOFFilePath(filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.file = f
	return nil
}

func (a *AOF) ParseManifest(data []string) error {
	if len(data) != 6 {
		return fmt.Errorf("Invalid manifest format")
	}
	if data[0] != "file" || data[2] != "seq" || data[4] != "type" {
		return fmt.Errorf("Invalid manifest format")
	}
	return nil
}


func (a *AOF) ReadManifestFile(filePath string) ([]Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to read AOF manifest: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	aofFiles := []Manifest{}
	for _, line := range lines {
		data := strings.Split(line, " ")
		if err := a.ParseManifest(data); err != nil {
			return nil, err
		}
		if data[5] == "i" {
			seq, _ := strconv.ParseInt(data[3], 10, 64)
			aofFiles = append(aofFiles, Manifest{
				AOFFilename: data[1],
				Sequence: seq,
				Type: "i",
			})
		}
	}
	sort.Slice(aofFiles, func(i, j int) bool {
		return aofFiles[i].Sequence < aofFiles[j].Sequence
	})
	return aofFiles, nil
}


// func (a *AOF) readManifestFile() (string, error) {
// 	f, err := os.Open(a.Manifest.Filename)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()

// 	scanner := bufio.NewScanner(f)

// 	var aofFile string
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		splitedLine := strings.Split(strings.TrimSpace(line), " ")
// 		filename := splitedLine[1]
// 		switch splitedLine[5] {
// 			case "i":
// 				aofFile = filename
// 		}

// 	}
// 	return aofFile, nil
// }


func manifestContent(a *AOF) string {
	return fmt.Sprintf("file %s.1.incr.aof seq %d type %s", a.AppendFilename, a.Manifest.Sequence, a.Manifest.Type)
}

func (a *AOF) WriteManifest() error {
	if a.AppendOnly != "yes" {
		return nil
	}
	f, err := os.OpenFile(a.GetAOFFilePath(a.ManifestFilename), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	content := manifestContent(a)
	_, err = f.WriteString(content)
	return err
}

