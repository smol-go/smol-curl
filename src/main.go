package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/url"
	"os"
	"path/filepath"
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

type formData struct {
	name     string
	value    string
	filename string
}

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
		{"-k", "<bool>", "Allow insecure server connections when using SSL"},
		{"-v", "<bool>", "Make the request more detailed"},
		{"-m", "<int>", "Maximum time allowed for the operation in seconds"},
		{"-u", "<string>", "Specify the user name and password for server authentication"},
		{"-o", "<string>", "Write the response body to the specified file"},
		{"-d", "<string>", "HTTP POST data"},
		{"-I", "<bool>", "Send HTTP HEAD request instead of GET"},
		{"-E", "<string>", "Specify the client certificate file for HTTPS"},
		{"-D", "<string>", "Write the response headers to the specified file"},
		{"-X", "<string>", "Specify custom request method"},
		{"-H", "<string[]>", "Pass custom header(s) to server"},
		{"-F", "<string[]>", "Specify HTTP multipart POST data"},
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

func createMultipartFormData(formParams []formData) (bytes.Buffer, string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, param := range formParams {
		var fw io.Writer
		var err error

		if param.filename != "" {
			fw, err = w.CreateFormFile(param.name, filepath.Base(param.filename))
			if err != nil {
				return bytes.Buffer{}, "", err
			}

			file, err := os.Open(param.filename)
			if err != nil {
				return bytes.Buffer{}, "", err
			}
			defer file.Close()

			_, err = io.Copy(fw, file)
			if err != nil {
				return bytes.Buffer{}, "", err
			}
		} else {
			fw, err = w.CreateFormField(param.name)
			if err != nil {
				return bytes.Buffer{}, "", err
			}

			_, err = fw.Write([]byte(param.value))
			if err != nil {
				return bytes.Buffer{}, "", err
			}
		}
	}

	w.Close()

	return b, w.FormDataContentType(), nil
}

func main() {
	userAgent := flag.String("a", "GolangHTTPClient/1.0", "Specify the User-Agent string")
	insecure := flag.Bool("k", false, "Allow insecure server connections when using SSL")
	verbose := flag.Bool("v", false, "Make the request more detailed")
	timeout := flag.Int("m", 0, "Maximum time allowed for the operation in seconds")
	userAuth := flag.String("u", "", "Specify the user name and password for server authentication")
	outputFile := flag.String("o", "", "Write the response body to the specified file")

	var postData stringSliceFlag
	flag.Var(&postData, "d", "HTTP POST data")

	certFile := flag.String("E", "", "Specify the client certificate file for HTTPS")
	headRequest := flag.Bool("I", false, "Send HTTP HEAD request instead of GET")
	headerFile := flag.String("D", "", "Write the response headers to the specified file")
	customMethod := flag.String("X", "", "Specify custom request method")

	var headers stringSliceFlag
	flag.Var(&headers, "H", "Pass custom header(s) to server")

	var formParams stringSliceFlag
	flag.Var(&formParams, "F", "Specify HTTP multipart POST data")

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
	method := "GET"
	if *customMethod != "" {
		method = strings.ToUpper(*customMethod)
	} else if *headRequest {
		method = "HEAD"
	} else if len(formParams) > 0 || len(postData) > 0 {
		method = "POST"
	}

	// Parse form data for -F flag
	var formDataSlice []formData
	for _, param := range formParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid form data: %s\n", param)
			return
		}
		name := parts[0]
		value := parts[1]

		if strings.HasPrefix(value, "@") {
			formDataSlice = append(formDataSlice, formData{name: name, filename: value[1:]})
		} else {
			formDataSlice = append(formDataSlice, formData{name: name, value: value})
		}
	}

	// Prepare request body
	var requestBody bytes.Buffer
	var contentType string
	if len(formDataSlice) > 0 {
		// Handle multipart form data (-F flag)
		requestBody, contentType, err = createMultipartFormData(formDataSlice)
		if err != nil {
			fmt.Printf("Error creating multipart form data: %s\n", err)
			return
		}
	} else if len(postData) > 0 {
		// Handle URL-encoded form data or raw data (-d flag)
		if strings.Contains(postData[0], "=") {
			// Treat as URL-encoded form data
			values := url.Values{}
			for _, data := range postData {
				parts := strings.SplitN(data, "=", 2)
				if len(parts) == 2 {
					values.Add(parts[0], parts[1])
				}
			}
			requestBody.WriteString(values.Encode())
			contentType = "application/x-www-form-urlencoded"
		} else {
			// Treat as raw data
			requestBody.WriteString(strings.Join(postData, "&"))
			contentType = "application/x-www-form-urlencoded"
		}
	}

	// Build the request headers
	requestHeaders := fmt.Sprintf(
		"%s / HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: */*\r\n", method, hostname, *userAgent)

	// Add Content-Type and Content-Length for POST data
	if requestBody.Len() > 0 {
		requestHeaders += fmt.Sprintf("Content-Type: %s\r\n", contentType)
		requestHeaders += fmt.Sprintf("Content-Length: %d\r\n", requestBody.Len())
	}

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
		fmt.Println("Failed to send request headers:", err)
		return
	}

	// Send the request body if present
	if requestBody.Len() > 0 {
		_, err = conn.Write(requestBody.Bytes())
		if err != nil {
			fmt.Println("Failed to send request body:", err)
			return
		}
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

	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(response.String())
		if err != nil {
			fmt.Printf("Failed to write response to file: %v\n", err)
		}

		fmt.Printf("Response saved to file: %s\n", *outputFile)
	}
}
