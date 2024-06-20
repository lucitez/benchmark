package benchmark

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lucitez/benchmark/crawler"
)

func benchmarkWebsite(url string, onUrl func(Performance)) {
	logger.Printf("Benchmarking %s...\n", url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	start := time.Now()

	urls := make(chan string)
	pch := make(chan Performance)

	go crawler.Crawl(url, 0, urls, &sync.Map{})
	go processUrls(urls, pch)

	performances := []Performance{}

	for performance := range pch {
		onUrl(performance)
		performances = append(performances, performance)
	}

	sort.Slice(performances, func(i, j int) bool {
		return performances[i].Latency > performances[j].Latency
	})

	for _, ping := range performances {
		logger.Printf("%+v\n", ping)
	}

	logger.Printf("Executed in %d millis\n", time.Since(start).Milliseconds())
}

type Performance struct {
	Url     string `json:"url"`
	Latency int64  `json:"latency"`
}

var logger = log.Default()

// benchmark requests the url 10 times, takes the average latency, returns a ping with that latency.
// bottleneck is here, this whole program is only as fast as the slowest crawled url.
// in UI, we should show progress instead of blocking while we wait for all urls.
func benchmarkURL(url string) Performance {
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

	return Performance{
		url,
		int64(totalLatencyMillis / 10),
	}
}

// as the crawler sends urls in the url chan, send them to the benchmarker.
// once each url has been benchmarked, close the performance chan.
func processUrls(urls <-chan string, pc chan<- Performance) {
	wg := sync.WaitGroup{}
	for url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			pc <- benchmarkURL(u)
		}(url)
	}
	wg.Wait()
	close(pc)
}
