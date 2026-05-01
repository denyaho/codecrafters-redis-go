package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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

func _splitStreamID(id string) (int64, int64, error) {
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

func _validateStreamID(current_id, prev_id string) (bool, error) {
	msInt, sqInt, err := _splitStreamID(current_id)
	if err != nil {
		return false, err
	}
	if msInt == 0 && sqInt == 0 {
		return false, fmt.Errorf("The ID specified in XADD must be greater than 0-0")
	}

	if prev_id != ""{
		prev_msInt, prev_sqInt, err := _splitStreamID(prev_id)
		if err != nil {
			return false, fmt.Errorf("invalid previous ID format: %v", err)
		}
		if msInt < prev_msInt || (msInt == prev_msInt && sqInt <= prev_sqInt) {
			return false, fmt.Errorf("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}
	return true, nil
}

func _resolveStreamID(rawID, prevID string, st *store.ExpireMap) (string, error) {
	if rawID == "*"{
		msInt := time.Now().UnixNano() / int64(time.Millisecond)
		rawID = strconv.FormatInt(msInt, 10) + "-*"
	}
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
			if msInt == 0{
				parts[1] = "1"
			} else {
				parts[1] = "0"
			}
		} else {
			prev_msInt, prev_sqInt, err := _splitStreamID(prevID)
			if err != nil {
				return "", fmt.Errorf("invalid previous ID format: %v", err)
			}
			if msInt == prev_msInt {
				parts[1] = strconv.FormatInt(prev_sqInt+1, 10)
			} else if msInt != prev_msInt {
				if msInt == 0{
					parts[1] = "1"
				} else {
					parts[1] = "0"
				}
			}
		}
		rawID = parts[0] + "-" + parts[1]
	}
	return rawID, nil
}

func _getLastStream(st *store.ExpireMap, key string) ([]StreamEntry) {
	var stream []StreamEntry

	value, ok := st.Get(key)
	if ok {
		if existingStream, exist := value.([]StreamEntry); exist {
			stream = existingStream
		}
	}
	return stream
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
	prevID := ""
	stream := _getLastStream(st, key)
	if len(stream) > 0 {
		prevID = stream[len(stream)-1].ID
	}

	entryID, err := _resolveStreamID(entryID, prevID, st)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	_, err = _validateStreamID(entryID, prevID)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	newEntry.ID = entryID
	stream = append(stream, newEntry)
	st.Set(key, stream, 0)
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n",len(entryID), entryID))
}

func _normalizeStreamID(id string) (string, error) {
	parts := strings.Split(id, "-")
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid stream ID format")
	}
	_, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid milliseconds in stream ID: %v", err)
	}
	if len(parts) == 1 {
		return parts[0] + "-*",nil
	}
	if parts[1] != "*" {
		return id, nil
	}
	_, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid sequence number in stream ID: %v", err)
	}
	return id, nil
}

func _getIndexOfStreamID(stream []StreamEntry, targetID string, isStart bool) int {
	target_ms, target_sq, _ := _splitStreamID(targetID)
	for idx, entry := range stream {
		entry_ms, entry_sq, _ := _splitStreamID(entry.ID)
		if entry_ms == target_ms && entry_sq == target_sq {
			return idx
		}
		if target_ms < entry_ms || (target_ms == entry_ms && target_sq < entry_sq) {
			if isStart {
				return idx
			}
			return idx -1
		}
	}
	return len(stream) - 1
}

func _resolveRangeID(rawID string, isStart bool) (string, error) {

	if rawID == "-" {
		return "0-0", nil
	}
	if rawID == "+" {
		return fmt.Sprintf("%d-%d", int64(math.MaxInt64), int64(math.MaxInt64)), nil
	}
	normID, err := _normalizeStreamID(rawID)
	if err != nil {
		return "", err
	}
	parts := strings.Split(normID, "-")
	if parts[1] == "*" {
		if isStart {
			return parts[0] + "-0", nil
		}
		return fmt.Sprintf("%s-%d",parts[0],int64(math.MaxInt64)), nil
	}
	return normID, nil	
}



func handleXRange(st *store.ExpireMap, args []string) []byte {
	if len(args) != 4{
		return []byte("-ERR wrong number of arguments for 'XRANGE' command\r\n")
	}
	key := args[1]
	startID, err := _resolveRangeID(args[2], true)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}
	endID, err := _resolveRangeID(args[3], false)
	if err != nil {
		return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
	}

	val, _ := st.Get(key)
	stream, ok := val.([]StreamEntry)
	if !ok {
		return []byte("-ERR no such key\r\n")
	}
	stream_matched := []StreamEntry{}

	start_idx := _getIndexOfStreamID(stream, startID, true)
	end_idx := _getIndexOfStreamID(stream, endID, false)
	if start_idx <= end_idx {
		stream_matched = stream[start_idx:end_idx+1]
	}
	fmt.Printf("will reply%v", stream_matched)
	response := []byte(fmt.Sprintf("*%d\r\n", len(stream_matched)))
	for i := 0; i < len(stream_matched); i++ {
		word := []byte(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(stream_matched[i].ID), stream_matched[i].ID))
		field_header := []byte(fmt.Sprintf("*%d\r\n", len(stream_matched[i].value) * 2))
		word = append(word, field_header...)
		for field, value := range stream_matched[i].value {
			field_word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(field), field))
			value_word := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
			word = append(word, field_word...)
			word = append(word, value_word...)
		}
		response = append(response, word...)
	}
	return response
}