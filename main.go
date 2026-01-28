package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	URL   string
	Data  string
	Error error
	Size  int
}

func main() {
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/xml",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
		"https://httpbin.org/user-agent",
	}
	sequentialScrape(urls)
	concurrentScrape(urls)
}

func sequentialScrape(urls []string) {
	start := time.Now()

	for _, url := range urls {
		_, size, err := fetchURL(url)
		if err != nil {
			fmt.Printf("Error fetchin %s: %v\n", url, err)
		}
		fmt.Printf("%s - %d bytes\n", url, size)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %v time\n", elapsed)
}

func concurrentScrape(urls []string) {
	start := time.Now()

	results := make(chan Result, len(urls))

	var wg sync.WaitGroup

	for _, url := range urls {

		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			data, size, err := fetchURL(u)
			results <- Result{
				URL:   u,
				Data:  data,
				Size:  size,
				Error: err,
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.Error != nil {
			fmt.Printf("Error fetchin %s: %v\n", res.URL, res.Error)
		}
		fmt.Printf("%s - %d bytes\n", res.URL, res.Size)
	}
	fmt.Printf("Concurent took %v time\n", time.Since(start))
}

func fetchURL(url string) (string, int, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(body)

	return string(body[:n]), n, nil
}
