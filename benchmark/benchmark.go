package benchmark

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lucitez/benchmark/crawler"
	"golang.org/x/sync/errgroup"
)

var MAX_DEPTH = 3

func benchmarkWebsite(rootURL string, urlOut chan<- string, benchmarkOut chan<- Benchmark) {
	logger.Printf("Benchmarking %s...\n", rootURL)

	start := time.Now()

	cr := crawler.New(rootURL, MAX_DEPTH)

	visitedIn := make(chan string)

	go cr.Crawl(visitedIn)

	urls := []string{}

	for visitedURL := range visitedIn {
		urls = append(urls, visitedURL)
		urlOut <- visitedURL
	}
	close(urlOut)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		benchmarkURLs(urls, benchmarkOut)
		wg.Done()
	}()

	wg.Wait()

	logger.Printf("Executed in %d millis\n", time.Since(start).Milliseconds())
}

type Benchmark struct {
	Url     string `json:"url"`
	Latency int64  `json:"latency"`
	Size    int64  `json:"size"`
}

var logger = log.Default()

// benchmark requests the url 10 times, takes the average latency, returns a ping with that latency.
// bottleneck is here, this whole program is only as fast as the slowest crawled url.
// in UI, we should show progress instead of blocking while we wait for all urls.
func benchmarkURL(url string) Benchmark {
	latencyChan := make(chan int64, 10)

	// TODO handle non 200 responses, errors, and timeouts
	// we are possibly skewing by returning early, not to mention introducing
	// a deadlock since we aren't sending to the chan on error
	for i := 0; i < 10; i++ {
		go func() {
			start := time.Now()

			_, err := http.Get(url)

			if err != nil {
				return
			}

			latencyChan <- time.Since(start).Milliseconds()
		}()
	}

	var totalLatencyMillis int64 = 0
	for i := 0; i < 10; i++ {
		totalLatencyMillis += <-latencyChan
	}

	return Benchmark{
		url,
		int64(totalLatencyMillis / 10),
		0,
	}
}

// benchmarkURLs ranges over the crawled urls and calls benchmarkURL on each.
// limit concurrency here because benchmarkURL spins up 10 goroutines each
// which can overload the network.
// each benchmark is sent to benchmarkOut as they come in
func benchmarkURLs(urls []string, benchmarkOut chan<- Benchmark) {
	g := new(errgroup.Group)
	g.SetLimit(20)
	for _, url := range urls {
		url := url
		g.Go(func() error {
			benchmarkOut <- benchmarkURL(url)
			return nil
		})
	}
	g.Wait()
	close(benchmarkOut)
}
