package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const MAXDATASIZE = 10000

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

type stringSliceFlag []string

func writeHeadersToFile(filename, headers string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(headers)
	return err
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func printFlagTable() {
	logo := `
███████╗███╗   ███╗ ██████╗ ██╗       ██████╗██╗   ██╗██████╗ ██╗     
██╔════╝████╗ ████║██╔═══██╗██║      ██╔════╝██║   ██║██╔══██╗██║     
███████╗██╔████╔██║██║   ██║██║█████╗██║     ██║   ██║██████╔╝██║     
╚════██║██║╚██╔╝██║██║   ██║██║╚════╝██║     ██║   ██║██╔══██╗██║     
███████║██║ ╚═╝ ██║╚██████╔╝███████╗ ╚██████╗╚██████╔╝██║  ██║███████╗
╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
`

	fmt.Print(colorCyan, logo, colorReset)

	fmt.Println(colorYellow, "Usage: ./main [options] <hostname>", colorReset)
	fmt.Println(colorGreen, "Options:", colorReset)

	table := [][]string{
		{"Flag", "Type", "Description"},
		{"-a", "<string>", "Specify the User-Agent string"},
		{"-E", "<string>", "Specify the client certificate file for HTTPS"},
		{"-I", "<bool>", "Send HTTP HEAD request instead of GET"},
		{"-k", "<bool>", "Allow insecure server connections when using SSL"},
		{"-v", "<bool>", "Make the request more detailed"},
		{"-m", "<int>", "Maximum time allowed for the operation in seconds"},
		{"-u", "<string>", "Specify the user name and password for server authentication"},
		{"-D", "<string>", "Write the response headers to the specified file"},
		{"-X", "<string>", "Specify custom request method"},
		{"-H", "<string>", "Pass custom header(s) to server"},
		{"--cookie", "<string>", "Send the specified cookies with the request"},
		{"--connect-timeout", "<int>", "Maximum time allowed for connection"},
	}

	columnWidths := []int{20, 10, 50}
	for i, row := range table {
		for j, cell := range row {
			if i == 0 {
				fmt.Print(colorPurple, padRight(cell, columnWidths[j]), colorReset)
			} else {
				switch j {
				case 0:
					fmt.Print(colorBlue, padRight(cell, columnWidths[j]), colorReset)
				case 1:
					fmt.Print(colorRed, padRight(cell, columnWidths[j]), colorReset)
				default:
					fmt.Print(padRight(cell, columnWidths[j]))
				}
			}
			fmt.Print(" | ")
		}
		fmt.Println()
		if i == 0 {
			fmt.Println(strings.Repeat("-", 85))
		}
	}
}

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	userAgent := flag.String("a", "GolangHTTPClient/1.0", "Specify the User-Agent string")
	certFile := flag.String("E", "", "Specify the client certificate file for HTTPS")
	headRequest := flag.Bool("I", false, "Send HTTP HEAD request instead of GET")
	insecure := flag.Bool("k", false, "Allow insecure server connections when using SSL")
	verbose := flag.Bool("v", false, "Make the request more detailed")
	timeout := flag.Int("m", 0, "Maximum time allowed for the operation in seconds")
	userAuth := flag.String("u", "", "Specify the user name and password for server authentication")
	headerFile := flag.String("D", "", "Write the response headers to the specified file")
	customMethod := flag.String("X", "", "Specify custom request method")
	var headers stringSliceFlag
	flag.Var(&headers, "H", "Pass custom header(s) to server")
	cookies := flag.String("cookie", "", "Send the specified cookies with the request")
	connectTimeout := flag.Int("connect-timeout", 0, "Maximum time allowed for the connection to be established in seconds")

	// Custom usage message
	flag.Usage = printFlagTable

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
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

	// Determine the HTTP method
	var method string
	switch {
	case *customMethod != "":
		method = strings.ToUpper(*customMethod)
	case *headRequest:
		method = "HEAD"
	default:
		method = "GET"
	}

	// Build the request headers
	requestHeaders := fmt.Sprintf(
		"%s / HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: */*\r\n", method, hostname, *userAgent)

	// Add cookies if provided
	if *cookies != "" {
		requestHeaders += fmt.Sprintf("Cookie: %s\r\n", *cookies)
	}

	// Add custom headers
	for _, header := range headers {
		requestHeaders += fmt.Sprintf("%s\r\n", header)
	}

	// Add authentication header if -u flag is used
	if *userAuth != "" {
		authHeader := fmt.Sprintf("Authorization: Basic %s", base64.StdEncoding.EncodeToString([]byte(*userAuth)))
		requestHeaders += fmt.Sprintf("%s\r\n", authHeader)
	}

	// Close the connection
	requestHeaders += "Connection: close\r\n\r\n"

	if *verbose {
		fmt.Println("Sending request:")
		fmt.Println(requestHeaders)
	}

	_, err = conn.Write([]byte(requestHeaders))
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}

	var response strings.Builder
	recvBuff := make([]byte, MAXDATASIZE)
	headersCaptured := false
	for {
		bytesRecvd, err := conn.Read(recvBuff)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading response:", err)
			}
			break
		}
		response.Write(recvBuff[:bytesRecvd])

		if !headersCaptured && *headerFile != "" {
			headersEnd := strings.Index(response.String(), "\r\n\r\n")
			if headersEnd != -1 {
				headersCaptured = true
				headers := response.String()[:headersEnd+4]

				err := writeHeadersToFile(*headerFile, headers)
				if err != nil {
					fmt.Printf("Failed to write headers to file: %s\n", err)
				} else if *verbose {
					fmt.Printf("Headers written to %s\n", *headerFile)
				}
			}
		}

		if *headRequest || method == "HEAD" {
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
