package handler

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"strings"
	"strconv"
)

type StreamEntry struct {
	ID string
	value map[string]string
}

func handleType(st *store.ExpireMap, args []string) []byte {
	if len(args) < 2 {
		return []byte("-ERR wrong number of arguments for 'TYPE' command\r\n")
	}
	value, ok := st.Get(args[1])
	if !ok {
		return []byte("+none\r\n")
	}

	switch value.(type) {	
	case string:
		return []byte("+string\r\n")
	case []string:
		return []byte("+list\r\n")
	case []StreamEntry:
		return []byte("+stream\r\n")
	default:
		return []byte("+none\r\n")
	}
}

func splitStreamID(id string) (int64, int64, error) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid stream ID format")
	}
	ms, sq := parts[0], parts[1]
	msInt, err := strconv.ParseInt(ms, 10, 64)	
	if err != nil {
		return 0, 0, fmt.Errorf("invalid milliseconds in stream ID: %v", err)
	}
	sqInt, err := strconv.ParseInt(sq, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sequence number in stream ID: %v", err)
	}
	return msInt, sqInt, nil
}

func validateStreamID(st *store.ExpireMap, key string, StreamID string) (bool, error) {
	msInt, sqInt, err := splitStreamID(StreamID)
	if err != nil {
		return false, err
	}

	stream, _ := st.Get(key)
	if stream == nil || stream.([]StreamEntry) == nil {
		if msInt == 0 && sqInt == 0 {
			return false, fmt.Errorf("The ID specified in XADD must be greater than 0-0")
		}
	}
	entries := stream.([]StreamEntry)
	lastStream := entries[len(entries)-1]

	prev_msInt, prev_sqInt, _ := splitStreamID(lastStream.ID)
	if msInt < prev_msInt || (msInt == prev_msInt && sqInt <= prev_sqInt) {
		return false, fmt.Errorf("The ID specified in XADD is equal or smaller than the target stream top item")
	}
	return true, nil
}

func handleXAdd(st *store.ExpireMap, args []string) []byte {
	if len(args) < 4 || len(args[3:])%2 != 0 {
		return []byte("-ERR wrong number of arguments for 'XADD' command\r\n")
	}
	key := args[1]

	entryID := args[2]
	pairs := make(map[string]string)
	for i := 3; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]
		pairs[field] = value
	}
	newEntry := StreamEntry{
		ID: entryID,
		value: pairs,
	}

	var stream []StreamEntry

	value, ok := st.Get(key)
	if ok {
		if existingStream, exist := value.([]StreamEntry); exist {
			stream = existingStream
		}
	}
	_, err := validateStreamID(st, key, entryID)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	stream = append(stream, newEntry)
	st.Set(key, stream, 0)
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n",len(entryID), entryID))
}
