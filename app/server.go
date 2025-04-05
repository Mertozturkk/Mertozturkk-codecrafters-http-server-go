package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
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
		if errors.Is(err, io.EOF) {
			return
		}

		req, err := NewRequest(buf[:n])
		if err != nil {
			fmt.Println("Error parsing request:", err)
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}
		if req.Method == "POST" {
			if strings.HasPrefix(req.Path, "/files") {
				directory := os.Args[2]
				fileName := strings.Split(req.Path, "/")[2]
				content := req.Body
				WriteFile(directory, fileName, content)
				conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			}

		}

		if req.Path == "/" {
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		} else if strings.HasPrefix(req.Path, "/echo") {
			header := GetHeaderValue(req.Headers, "Accept-Encoding")
			message := strings.Split(req.Path, "/")[2]

			if header == "gzip" {
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))

			} else {
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
			}

			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
		} else if strings.HasPrefix(req.Path, "/user-agent") {
			userAgent := req.Headers["User-Agent"]
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
		} else if strings.HasPrefix(req.Path, "/files") {
			fileName := strings.Split(req.Path, "/")[2]

			directory := os.Args[2]
			content, err := ReadFileFromFileName(directory, fileName)
			if err != nil {
				fmt.Println("Error reading file:", err)
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)))
		} else {
			fmt.Println("404 Not Found")
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

	}
}

type Response struct {
	Status  string
	Headers map[string]string
	Body    string
}

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
	fmt.Println("request.Headers", request.Headers)

	request.Body = strings.Join(lines[len(lines)-1:], "\r\n")
	return request, nil
}

func ReadFileFromFileName(directory, fileName string) ([]byte, error) {
	filePath := fmt.Sprintf("%s/%s", directory, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func WriteFile(directory, fileName, content string) {
	filepath := filepath.Join(directory, fileName)
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File %s created successfully\n", filepath)

}

func GetHeaderValue(headers map[string]string, key string) string {
	if value, ok := headers[key]; ok {
		return value
	}
	return ""
}
