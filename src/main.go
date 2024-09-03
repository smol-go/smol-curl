package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const MAXDATASIZE = 10000

func main() {
	userAgent := flag.String("a", "GolangHTTPClient/1.0", "Specify the User-Agent string")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: go run main.go [-a user-agent] <hostname>")
		return
	}

	urlToGet := args[0]
	hostname := strings.TrimPrefix(strings.TrimPrefix(urlToGet, "http://"), "https://")
	hostname = strings.Split(hostname, "/")[0]

	fmt.Printf("Fetching %s ...\n", hostname)

	addrs, err := net.LookupHost(hostname)
	if err != nil {
		fmt.Printf("LookupHost error: %s\n", err)
		return
	}

	var conn net.Conn
	for _, addr := range addrs {
		fmt.Println("Trying address:", addr)
		conn, err = net.DialTimeout("tcp", net.JoinHostPort(addr, "80"), 5*time.Second)
		if err != nil {
			fmt.Printf("Connection failed: %s\n", addr)
			fmt.Printf("Reason: %s\n", err)
			continue
		}
		fmt.Printf("Connected to %s\n", addr)
		break
	}

	if conn == nil {
		fmt.Println("Failed to connect to any address")
		return
	}
	defer conn.Close()

	sendBuff := fmt.Sprintf(
		"GET / HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: */*\r\n"+
			"Connection: close\r\n"+
			"\r\n", hostname, *userAgent)

	_, err = conn.Write([]byte(sendBuff))
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}

	var response strings.Builder
	recvBuff := make([]byte, MAXDATASIZE)
	for {
		bytesRecvd, err := conn.Read(recvBuff)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading response:", err)
			}
			break
		}
		response.Write(recvBuff[:bytesRecvd])
	}

	fmt.Println("Received:")
	fmt.Println(response.String())
	fmt.Println(strings.Repeat("-", 50))
}
