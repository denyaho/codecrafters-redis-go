package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
	"strconv"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

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
				reader.ReadString('\n')
				word, _ := reader.ReadString('\n')
				args[i] = strings.TrimSpace(word)
			}

			if strings.ToUpper(args[0]) == "ECHO" {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])))
			}else {
				reply_pong(conn)
			}
		}
		reply_pong(conn)
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
