package main

import (
	"fmt"
	"os"
	"github.com/codecrafters-io/redis-starter-go/internal/aof"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
)

func initAOF(config *Config, st *store.ExpireMap) (*aof.AOF, error) {
	aof := aof.NewAOF(config.CurrentDir, config.AppendOnly, config.AppendDirname, config.AppendFilename, config.AppendFsync)

	AOFdirectory := fmt.Sprintf("%s/%s", aof.Dir, aof.AppendDirname)

	if IsDir(AOFdirectory) && aof.AppendOnly == "yes" {
		err := AOFReplay(aof, st)
		if err != nil {
			return nil, err
		}
	}
	return aof, nil
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}


func AOFReplay(aof *aof.AOF, st *store.ExpireMap) error {
	aofDirectory := fmt.Sprintf("%s/%s", aof.Dir, aof.AppendDirname)
	manifestFile := aof.AppendFilename + ".manifest"
	manifestPath := fmt.Sprintf("%s/%s", aofDirectory, manifestFile)
	aofFiles, err := aof.ReadManifestFile(manifestPath)
	if err != nil {
		return err
	}
	fmt.Printf("AOF files to replay: %v\n", aofFiles)
	for _, aofFile := range aofFiles {
		
		aofFilePath := fmt.Sprintf("%s/%s", aofDirectory, aofFile.AOFFilename)
		fmt.Printf("Replaying AOF file: %s\n", aofFilePath)
		f, err := os.Open(aofFilePath)
		if err != nil {
			return fmt.Errorf("Failed to open AOF file: %v", err)
		}
		defer f.Close()

		err = handler.ApplyForReplay(st, f)
		if err != nil {
			return err
		}
	}
	return nil
}