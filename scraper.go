package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ScraperConfig struct {
	MaxWorkers int
	Timeout    time.Duration
	UserAgent  string
	MaxRetries int
	RateLimit  time.Duration
}

type Result struct {
	URL       string
	Data      string
	Error     error
	Size      int
	Duration  time.Duration
	Status    int
	WorkerId  int
	Retries   int
	Timestamp time.Time
}

type Stats struct {
	TotalRequests  int
	SuccessRequest int
	FailedRequests int
	TotalBytes     int64
	TotalDuration  time.Duration
}

type Scraper struct {
	config ScraperConfig
	client *http.Client
	stats  Stats
	mu     sync.Mutex
}

func NewScraper(config ScraperConfig) *Scraper {
	if config.MaxWorkers == 0 {
		config.MaxWorkers = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Scraper{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		stats: Stats{},
	}
}

func (s *Scraper) ScrapeURLs(urls []string) []Result {
	ctx := context.Background()
	return s.ScrapeURLsWithContext(ctx, urls)
}

func (s *Scraper) ScrapeURLsWithContext(ctx context.Context, urls []string) []Result {
	jobChan := make(chan string, len(urls))
	resultChan := make(chan Result, len(urls))

	var wg sync.WaitGroup

	for i := range s.config.MaxWorkers {
		wg.Go(func() {
			workerRes := s.worker(ctx, i, jobChan)
			for res := range workerRes {
				resultChan <- res
			}
		})
	}

	go func() {
		for _, url := range urls {
			jobChan <- url
		}
		close(jobChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]Result, 0, len(urls))
	for result := range resultChan {
		results = append(results, result)

		s.mu.Lock()
		s.stats.TotalRequests++
		if result.Error != nil {
			s.stats.FailedRequests++
		} else {
			s.stats.SuccessRequest++
			s.stats.TotalBytes += int64(result.Size)
			s.stats.TotalDuration += result.Duration
		}
		s.mu.Unlock()
	}
	return results

}

func (s *Scraper) worker(ctx context.Context, id int, jobs <-chan string) <-chan Result {
	results := make(chan Result)
	go func() {
		defer close(results)
		for url := range jobs {
			if s.config.RateLimit > 0 {
				time.Sleep(s.config.RateLimit)
			}

			var result Result
			for attempt := range s.config.MaxRetries {
				result = s.fetchURL(url)
				result.WorkerId = id
				result.Retries = attempt
				result.Timestamp = time.Now()
				if result.Error == nil {
					break
				}

				if attempt < s.config.MaxRetries {
					backoff := time.Duration(attempt+1) * time.Second
					select {
					case <-time.After(backoff):
						continue
					case <-ctx.Done():
						result.Error = ctx.Err()
						break
					}
				}
			}
			results <- result
		}
	}()
	return results
}

func (s *Scraper) fetchURL(url string) Result {
	start := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Result{URL: url, Error: err, Duration: time.Since(start)}
	}
	if s.config.UserAgent != "" {
		req.Header.Set("User-Agent", s.config.UserAgent)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err, Duration: time.Since(start)}
	}
	defer resp.Body.Close()

	body := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(body)
	return Result{
		URL:      url,
		Data:     string(body[:n]),
		Size:     n,
		Duration: time.Since(start),
		Status:   resp.StatusCode,
	}
}

func (s *Scraper) GetStats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprint("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
