package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
	"strconv"
	"io"
	"time"
	"sync"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

type Item struct {
	value string
	expireAt int64
}

type ExpireMap struct {
	data map[string]Item
	mu sync.RWMutex
}

func (m *ExpireMap) Set(key, value string, expireAt time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp int64
	if expireAt == 0 {
		exp = 0
	}else {
		exp = time.Now().Add(expireAt).Unix()
	}

	m.data[key] = Item{
		value: value,
		expireAt: exp,
	}
}

func (m *ExpireMap) Get(key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.data[key]
	if !ok {
		return "", false
	}
	if item.expireAt != 0 && item.expireAt < time.Now().Unix() {
		return "", false
	}
	return item.value, true
}

func (m *ExpireMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok :=m.data[key]; ok {
		delete(m.data, key)
	}
}


func reply_pong(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed")
			return
		}
		conn.Write([]byte("+PONG\r\n"))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	emap := &ExpireMap{
		data: make(map[string]Item),
		mu: sync.RWMutex{},
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed to read from connection: ", err.Error())
			return
		}
		if message[0] == '*' {
			countstr := strings.TrimSpace(message[1:])
			count, _ := strconv.Atoi(countstr)
			
			args := make([]string, count)
			for i := 0; i < count; i++ {
				line, _ := reader.ReadString('\n')
				if line[0] != '$' {
					continue;
				}
				wordcounter := strings.TrimSpace(line[1:])
				size, _ := strconv.Atoi(wordcounter)
				buf := make([]byte, size)
				io.ReadFull(reader, buf)
				reader.ReadString('\n') // read the trailing \r\n
				args[i] = string(buf)
			}

			if strings.ToUpper(args[0]) == "ECHO" {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])))
			}else if strings.ToUpper(args[0]) == "SET" {
				if len(args) == 4 && strings.ToUpper(args[3]) == "PX" {
					expireMs, _ := strconv.Atoi(args[4])
					expireAt := time.Duration(expireMs) * time.Millisecond
					emap.Set(args[1], args[2], expireAt)
				}else if len(args) == 4 && strings.ToUpper(args[3]) == "EX" {
					expireS, _ := strconv.Atoi(args[4])
					expireAt := time.Duration(expireS) * time.Second
					emap.Set(args[1], args[2], expireAt)
				}else{
					emap.Set(args[1], args[2], 0)
				}
				conn.Write([]byte("+OK\r\n"))
			}else if strings.ToUpper(args[0]) == "GET"{
				if val, ok := emap.Get(args[1]); ok {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
				} else {
					emap.Delete(args[1])
					conn.Write([]byte("$-1\r\n"))
				}
			}else {
				conn.Write([]byte("+PONG\r\n"))
			}
		}
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}
