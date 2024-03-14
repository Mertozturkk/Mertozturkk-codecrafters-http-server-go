package main

import (
	"errors"
	"fmt"
	"io"
	"log"

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
		// fmt.Println("Received int", n)
		receiveMessage := string(buf[:n])
		log.Printf("Received Data %s", receiveMessage)
		if errors.Is(err, io.EOF) {
			return
		}
		conn.Write([]byte("HTTP/1.1 200 OK" + "\r\n\r\n"))

	}
}
