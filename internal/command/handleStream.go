package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
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

func validateStreamID(current_id, prev_id string) (bool, error) {
	msInt, sqInt, err := splitStreamID(current_id)
	if err != nil {
		return false, err
	}
	if msInt == 0 && sqInt == 0 {
		return false, fmt.Errorf("The ID specified in XADD must be greater than 0-0")
	}

	if prev_id != ""{
		prev_msInt, prev_sqInt, err := splitStreamID(prev_id)
		if err != nil {
			return false, fmt.Errorf("invalid previous ID format: %v", err)
		}
		if msInt < prev_msInt || (msInt == prev_msInt && sqInt <= prev_sqInt) {
			return false, fmt.Errorf("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}
	return true, nil
}

func resolveStreamID(rawID, prevID string, st *store.ExpireMap) (string, error) {
	parts := strings.Split(rawID, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid stream ID format")
	}
	if parts[1] == "*"{
		msInt, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid milliseconds in stream ID: %v", err)
		}
		if prevID == ""{
			parts[1] = "0"
		} else {
			prev_msInt, prev_sqInt, err := splitStreamID(prevID)
			if err != nil {
				return "", fmt.Errorf("invalid previous ID format: %v", err)
			}
			if msInt == prev_msInt {
				parts[1] = strconv.FormatInt(prev_sqInt+1, 10)
			} else if msInt != prev_msInt {
				parts[1] = "0"
			}
		}
		rawID = parts[0] + "-" + parts[1]
	}
	return rawID, nil
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
	prevID := ""
	if ok {
		if existingStream, exist := value.([]StreamEntry); exist {
			stream = existingStream
			prevID = stream[len(stream)-1].ID
		}
	}
	entryID, err := resolveStreamID(entryID, prevID, st)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	fmt.Printf("Resolved stream ID: %s\n", entryID)
	_, err = validateStreamID(entryID, prevID)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	newEntry.ID = entryID
	stream = append(stream, newEntry)
	st.Set(key, stream, 0)
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n",len(entryID), entryID))
}
