package resp

import (
	"encoding/hex"
	"fmt"
)

func BuildSimpleString(s string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

func BuildError(s string) []byte {
	return []byte(fmt.Sprintf("-%s", s))
}

func BuildBulkStrings(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func BuildNullBulkString() []byte {
	return []byte("$-1\r\n")
}

func BuildNullArray() []byte {
	return []byte("*-1\r\n")
}

func BuildInteger(i int) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", i))
}

func BuildArray(arr []string) []byte {
	if len(arr) == 0 {
		return []byte("*0\r\n")
	}
	response := []byte(fmt.Sprintf("*%d\r\n", len(arr)))
	for _, elem := range arr {
		response = append(response, BuildBulkStrings(elem)...)
	}
	return response
}

var RDBcontent, _ = hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")

func BuildRDB() []byte {
	header := []byte(fmt.Sprintf("$%d\r\n", len(RDBcontent)))
	return append(header, RDBcontent...)	
}
