package resp

import (
	"fmt"
)

func BuildSimpleString(s string) []byte {
	return []byte("+" + s + "\r\n")
}

func BuildError(s string) []byte {
	return []byte("-" + s + "\r\n")
}

func BuildBulkStrings(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func BuildNullBulkString() []byte {
	return []byte("$-1\r\n")
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
