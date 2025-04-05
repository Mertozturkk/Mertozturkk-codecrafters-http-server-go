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

		req, err := NewRequest(buf[:n])
		if err != nil {
			fmt.Println("Error parsing request:", err)
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}
		fmt.Println("REEEQUEST:", req)

		if req.Path == "/" {
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		} else if strings.HasPrefix(req.Path, "/echo") {
			message := strings.Split(req.Path, "/")[2]
			fmt.Println("message", message)
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
		} else if strings.HasPrefix(req.Path, "/user-agent") {
			userAgent := req.Headers["User-Agent"]
			fmt.Println("User-Agent>>>>>>>>>>>>>>>>:", userAgent)
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
		} else {
			fmt.Println("404 Not Found")
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

		// RequestPathSplit(receiveMessage, conn)

	}
}

// func RequestPathSplit(receiveMessage string, conn net.Conn) {
// 	splitedMessage := strings.Split(receiveMessage, "\r\n")[0]
// 	fmt.Println("splitedMessage", splitedMessage)
// 	targetPath := strings.Split(splitedMessage, " ")[1]
// 	fmt.Println("targetPath", targetPath)

// 	if targetPath == "/" {
// 		conn.Write([]byte("HTTP/1.1 200 OK" + "\r\n\r\n"))
// 	} else if strings.HasPrefix(targetPath, "/echo") {
// 		message := strings.Split(targetPath, "/")[2]
// 		fmt.Println("message", message)
// 		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
// 	} else {
// 		fmt.Println("404 Not Found")
// 		conn.Write([]byte("HTTP/1.1 404 Not Found" + "\r\n\r\n"))

// 	}

// }

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func NewRequest(b []byte) (*Request, error) {
	request := &Request{
		Headers: make(map[string]string),
	}

	lines := strings.Split(string(b), "\r\n")
	if len(lines) < 1 {
		return nil, errors.New("invalid request")
	}

	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		return nil, errors.New("invalid request line")
	}

	request.Method = requestLine[0]
	request.Path = requestLine[1]

	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		header := strings.Split(line, ": ")
		if len(header) != 2 {
			return nil, errors.New("invalid header")
		}
		request.Headers[header[0]] = header[1]
	}

	return request, nil
}
