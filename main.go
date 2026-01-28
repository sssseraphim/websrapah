package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/xml",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
		"https://httpbin.org/user-agent",
	}

	start := time.Now()

	for _, url := range urls {
		data, err := fetchURL(url)
		if err != nil {
			fmt.Printf("Error fetchin %s: %v\n", url, err)
		}
		fmt.Printf("%s - %d bytes\n", url, len(data))
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %v time", elapsed)
}

func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(body)

	return string(body[:n]), nil
}
