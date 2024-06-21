package benchmark

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lucitez/benchmark/crawler"
)

func benchmarkWebsite(rootUrl string, urlOut chan<- string, benchmarkOut chan<- Benchmark) {
	logger.Printf("Benchmarking %s...\n", rootUrl)

	if !strings.HasPrefix(rootUrl, "http://") && !strings.HasPrefix(rootUrl, "https://") {
		rootUrl = "https://" + rootUrl
	}

	start := time.Now()

	cr := crawler.New(rootUrl, 2)

	visited := make(chan string)

	go cr.Crawl(visited)

	urls := []string{}

	for visitedUrl := range visited {
		urls = append(urls, visitedUrl)
		urlOut <- visitedUrl
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

// as the crawler sends urls in the url chan, send them to the benchmarker.
// once each url has been benchmarked, close the performance chan.
func benchmarkURLs(urls []string, pc chan<- Benchmark) {
	wg := sync.WaitGroup{}
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			pc <- benchmarkURL(u)
		}(url)
	}
	wg.Wait()
	close(pc)
}
