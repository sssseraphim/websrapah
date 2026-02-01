package main

import (
	"fmt"
	"time"
)

func main() {
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/xml",
		"https://httpbin.org/robots.txt",
		"https://httpbin.org/headers",
		"https://httpbin.org/ip",
		"https://httpbin.org/user-agent",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/500",
		"https://httpbin.org/status/429", // Rate limit
		"https://httpbin.org/redirect/1",
		"https://httpbin.org/redirect/2",
		"https://httpbin.org/relative-redirect/1",
	}

	config := ScraperConfig{
		MaxWorkers: 10,
		Timeout:    10 * time.Second,
		UserAgent:  "lilDroptop",
		MaxRetries: 2,
		RateLimit:  100 * time.Millisecond,
	}

	scraper := NewScraper(config)
	fmt.Println("Starting scrapin ", len(urls), "URLs...")
	fmt.Println("Workers:", config.MaxWorkers)
	fmt.Println("Timeout:", config.Timeout)

	start := time.Now()
	scraper.ScrapeURLs(urls)
	elapsed := time.Since(start)

	stats := scraper.GetStats()
	fmt.Println("Took time: ", elapsed)
	fmt.Printf("%+v\n", stats)

}
