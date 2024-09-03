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
	insecure := flag.Bool("k", false, "Allow insecure server connections when using SSL")
	verbose := flag.Bool("v", false, "Make the request more detailed")
	timeout := flag.Int("m", 0, "Maximum time allowed for the operation in seconds")
	connectTimeout := flag.Int("connect-timeout", 0, "Maximum time allowed for the connection to be established in seconds")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: ./main [options] <hostname>")
		return
	}

	urlToGet := args[0]
	hostname := strings.TrimPrefix(strings.TrimPrefix(urlToGet, "http://"), "https://")
	hostname = strings.Split(hostname, "/")[0]

	if *verbose {
		fmt.Printf("Fetching %s ...\n", hostname)
	}

	addrs, err := net.LookupHost(hostname)
	if err != nil {
		fmt.Printf("LookupHost error: %s\n", err)
		return
	}

	var conn net.Conn
	for _, addr := range addrs {
		if *verbose {
			fmt.Println("Trying address:", addr)
		}

		// Set up connection timeout
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		if *connectTimeout > 0 {
			dialer.Timeout = time.Duration(*connectTimeout) * time.Second
		}

		if *timeout > 0 {
			dialer.Deadline = time.Now().Add(time.Duration(*timeout) * time.Second)
		}

		if *certFile != "" || *insecure {
			// If a certificate file is provided, use TLS, or if -k is used
			config := &tls.Config{
				InsecureSkipVerify: *insecure, // This corresponds to the -k flag
			}

			if *certFile != "" {
				cert, err := tls.LoadX509KeyPair(*certFile, *certFile)
				if err != nil {
					fmt.Printf("Error loading client certificate: %s\n", err)
					return
				}
				config.Certificates = []tls.Certificate{cert}
			}

			conn, err = tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(addr, "443"), config)
			if err != nil {
				fmt.Printf("Connection failed: %s\n", addr)
				fmt.Printf("Reason: %s\n", err)
				continue
			}
		} else {
			// Use regular TCP connection if no certificate is provided and -k is not used
			conn, err = net.DialTimeout("tcp", net.JoinHostPort(addr, "80"), 5*time.Second)
			if err != nil {
				fmt.Printf("Connection failed: %s\n", addr)
				fmt.Printf("Reason: %s\n", err)
				continue
			}
		}
		if *verbose {
			fmt.Printf("Connected to %s\n", addr)
		}
		break
	}

	if conn == nil {
		fmt.Println("Failed to connect to any address")
		return
	}
	defer conn.Close()

	// Set a deadline for the entire operation if the -m flag is used
	if *timeout > 0 {
		conn.SetDeadline(time.Now().Add(time.Duration(*timeout) * time.Second))
	}

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

	if *verbose {
		fmt.Println("Sending request:")
		fmt.Println(sendBuff)
	}

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

	if *verbose {
		fmt.Println("Received response:")
	}
	fmt.Println(response.String())
	fmt.Println(strings.Repeat("-", 50))
}
