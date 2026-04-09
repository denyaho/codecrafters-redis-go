package resp

import (
	"fmt"
	"strings"
	"strconv"
	"io"
	"bufio"
)

func Parse(reader *bufio.Reader) ([]string, error)  {
	message, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read from connection: %v", err)
	}
	if message[0] != '*' {
		return nil, fmt.Errorf("expected * but got %c", message[0])
	}
	countstr := strings.TrimSpace(message[1:])
	count, _ := strconv.Atoi(countstr)
	args := make([]string, count)
	for i := 0; i < count; i++ {
		line, _ := reader.ReadString('\n')
		if line[0] != '$' {
			continue
		}
		wordcounter := strings.TrimSpace(line[1:])
		size, _ := strconv.Atoi(wordcounter)
		buf := make([]byte, size)
		io.ReadFull(reader, buf)
		reader.ReadString('\n')
		args[i] = string(buf)
	}

	return args, nil
}
