package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	URL      string
	Data     string
	Error    error
	Size     int
	WorkerId int
	Duration time.Duration
}

func main() {
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/xml",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
		"https://httpbin.org/user-agent",
	}
	fmt.Println("\n no limit")
	scrapeWithWorkers(urls, 0) // 0 = no limit

	fmt.Println("\n3 workers ")
	scrapeWithWorkers(urls, 3)

	fmt.Println("\n 5 workers ")
	scrapeWithWorkers(urls, 5)
}

func scrapeWithWorkers(urls []string, maxWorkers int) {
	start := time.Now()

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	if maxWorkers == 0 {
		maxWorkers = len(urls)
	}
	var wg sync.WaitGroup
	for i := range maxWorkers {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	success := 0
	failure := 0
	for result := range results {
		if result.Error != nil {
			fmt.Printf("Worker %v failed: %v %v\n", result.WorkerId, result.URL, result.Error)
			failure++
		} else {
			fmt.Printf("Worker %v succeded: %v %v\n", result.WorkerId, result.URL, result.Duration.Milliseconds())
			success++
		}
	}
	fmt.Printf("%v succeded, %v failed, took %v seconds", time.Since(start).Seconds())
}

func worker(id int, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range jobs {
		start := time.Now()
		data, size, err := fetchURL(url)
		duration := time.Since(start)

		results <- Result{
			WorkerId: id,
			URL:      url,
			Data:     data,
			Error:    err,
			Size:     size,
			Duration: duration,
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func fetchURL(url string) (string, int, error) {
	client := http.Client{
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
