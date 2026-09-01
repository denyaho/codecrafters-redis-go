package handler

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"errors"
	"os"
	"io"
)

func ApplyForReplay(st *store.ExpireMap, file *os.File) error {
	reader := bufio.NewReader(file)
	for {
		args, err := resp.Parse(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("Failed to parse AOF file: %v", err)
		}
		if len(args) == 0 {
			continue
		}
		command := strings.ToUpper(args[0])
		switch command {
		case "SET":
			_, _ = handleSet(st, args)
		case "RPUSH":
			_, _ = handleRpush(st, args)
		case "LPUSH":
			_, _ = handleLpush(st, args)
		case "LPOP":
			_, _ = handleLpop(st, args)
		case "BLPOP":
			_, _ = handleBLpop(st, args)
		case "XADD":
			_, _ = handleXAdd(st, args)
		case "ZREM":
			_, _ = handleZREM(st, args)
		case "GEOADD":
			_, _ = handleGEOADD(st, args)
		default:
			return fmt.Errorf("Unsupported command in AOF: %s", command)
		}
	}
	return nil

}