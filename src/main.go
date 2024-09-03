package main

import (
	"crypto/tls"
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
	certFile := flag.String("E", "", "Specify the client certificate file for HTTPS")
	headRequest := flag.Bool("I", false, "Send HTTP HEAD request instead of GET")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: ./main [-a user-agent] <hostname>")
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
		if *certFile != "" {
			// If a certificate file is provided, use TLS
			cert, err := tls.LoadX509KeyPair(*certFile, *certFile)
			if err != nil {
				fmt.Printf("Error loading client certificate: %s\n", err)
				return
			}
			config := &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true, // Note: This is not secure for production use
			}
			conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", net.JoinHostPort(addr, "443"), config)
			if err != nil {
				fmt.Printf("Connection failed: %s\n", addr)
				fmt.Printf("Reason: %s\n", err)
				continue
			}
		} else {
			// Use regular TCP connection if no certificate is provided
			conn, err = net.DialTimeout("tcp", net.JoinHostPort(addr, "80"), 5*time.Second)
			if err != nil {
				fmt.Printf("Connection failed: %s\n", addr)
				fmt.Printf("Reason: %s\n", err)
				continue
			}
		}
		fmt.Printf("Connected to %s\n", addr)
		break
	}

	if conn == nil {
		fmt.Println("Failed to connect to any address")
		return
	}
	defer conn.Close()

	method := "GET"
	if *headRequest {
		method = "HEAD"
	}

	sendBuff := fmt.Sprintf(
		"%s / HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: */*\r\n"+
			"Connection: close\r\n"+
			"\r\n", method, hostname, *userAgent)

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

		if *headRequest {
			// If it's a HEAD request, break after reading the headers
			if strings.Contains(response.String(), "\r\n\r\n") {
				break
			}
		}
	}

	fmt.Println("Received:")
	fmt.Println(response.String())
	fmt.Println(strings.Repeat("-", 50))
}
