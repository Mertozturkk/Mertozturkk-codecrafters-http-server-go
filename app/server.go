package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Server listening on ", l.Addr().String())

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error", err)
			continue
		}
		go HandleFunction(conn)
	}
}

func HandleFunction(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		//fmt.Println("Received int", n)
		receiveMessage := string(buf[:n])
		log.Printf("Received Data %s", receiveMessage)
		if errors.Is(err, io.EOF) {
			return
		}

		RequestPathSplit(receiveMessage, conn)

	}
}

func RequestPathSplit(receiveMessage string, conn net.Conn) {
	splitedMessage := strings.Split(receiveMessage, "\r\n")[0]
	fmt.Println("splitedMessage", splitedMessage)
	targetPath := strings.Split(splitedMessage, " ")[1]
	fmt.Println("targetPath", targetPath)

	if targetPath == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK" + "\r\n\r\n"))
	} else if strings.HasPrefix(targetPath, "/echo") {
		message := strings.Split(targetPath, "/")[2]
		fmt.Println("message", message)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else {
		fmt.Println("404 Not Found")
		conn.Write([]byte("HTTP/1.1 404 Not Found" + "\r\n\r\n"))

	}

}
