package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./main <url>")
		return
	}

	urlToGet := os.Args[1]
	fmt.Printf("Fetching %s ...\n", urlToGet)

	resp, err := http.Get(urlToGet)
	if err != nil {
		fmt.Printf("Failed to fetch URL: %s\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err)
		return
	}

	fmt.Println("Received:")
	fmt.Println(string(body))
	fmt.Println(strings.Repeat("-", 50))
}
