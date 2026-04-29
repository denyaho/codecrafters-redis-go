package resp

import (
	"fmt"
	"strings"
	"strconv"
	"io"
	"bufio"
)

func parseBulkString(reader *bufio.Reader) (string, error) {
	line, _ := reader.ReadString('\n')
	if line[0] != '$' {
		return "", fmt.Errorf("expected $ but got %c", line[0])
	}
	word_counter := strings.TrimSpace(line[1:])
	size, _ := strconv.Atoi(word_counter)
	buf := make([]byte, size)
	io.ReadFull(reader, buf)
	reader.ReadString('\n')
	return string(buf), nil
}

func parseArray(reader *bufio.Reader, count int ) ([]string, error) {
	args := make([]string, count)
	for i := 0; i < count; i++{
		bulk_string, err := parseBulkString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bulk string: %v", err)
		}
		args[i] = bulk_string
	}
	return args, nil
}

func Parse(reader *bufio.Reader) ([]string, error)  {
	message, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read from connection: %v", err)
	}
	if len(message) == 0 {
		return nil, err
	}
	switch message[0] {
		case '+':
			return []string{strings.TrimSpace(message[1:])}, nil
		case '*':
			countStr := strings.TrimSpace(message[1:])
			count, _ := strconv.Atoi(countStr)
			return parseArray(reader, count)
		default:
			return nil, fmt.Errorf("unexpected message type: %c", message[0])
	}

}
