package benchmarker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lucitez/benchmark/client"
	"golang.org/x/sync/errgroup"
)

type Benchmark struct {
	Url     string `json:"url"`
	Latency int64  `json:"latency"`
	Size    int64  `json:"size"`
}

type Benchmarker struct {
	Logger *log.Logger
	Client http.Client
}

func New() Benchmarker {
	return Benchmarker{
		Logger: log.Default(),
		Client: client.Http,
	}
}

// benchmarkWebsite ranges over the passed urls and calls benchmarkURL on each.
// limit concurrency here because benchmarkURL many goroutines for each url,
// which can end up throttling i/o
// each benchmark is sent to out as they come in
// TODO add error handling
func (b Benchmarker) BenchmarkWebsite(urls []string, out chan<- Benchmark) {
	b.Logger.Printf("Benchmarking...\n")

	start := time.Now()

	eg := errgroup.Group{}
	eg.SetLimit(30)
	for _, url := range urls {
		url := url
		eg.Go(func() error {
			benchmark, err := b.benchmarkURL(url)
			if err != nil {
				return err
			}
			out <- benchmark
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		b.Logger.Printf("Error while benchmarking URL: %v\n", err)
	}
	close(out)

	b.Logger.Printf("Executed in %d millis\n", time.Since(start).Milliseconds())
}

var NUM_REQUESTS int64 = 5

// benchmark requests the url NUM_REQUESTS times, takes the average latency, returns a ping with that latency.
// bottleneck is here, this whole program is only as fast as the slowest crawled url.
// in UI, we should show progress instead of blocking while we wait for all urls.
func (b Benchmarker) benchmarkURL(url string) (Benchmark, error) {
	latencyChan := make(chan int64, NUM_REQUESTS)
	sizeChan := make(chan int64, NUM_REQUESTS)

	// TODO handle non 200 responses, errors, and timeouts
	// we are possibly skewing by returning early, not to mention introducing
	// a deadlock since we aren't sending to the chan on error
	eg := errgroup.Group{}
	for i := 0; i < int(NUM_REQUESTS); i++ {
		eg.Go(func() error {
			start := time.Now()

			resp, err := b.Client.Get(url)

			if err != nil {
				return err
			}

			latencyChan <- time.Since(start).Milliseconds()

			written, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				return err
			}

			fmt.Printf("written: %v\n", written)

			sizeChan <- written
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return Benchmark{}, err
	}

	var totalLatencyMillis int64 = 0
	var totalSizeMillis int64 = 0
	for i := 0; i < int(NUM_REQUESTS); i++ {
		totalLatencyMillis += <-latencyChan
		totalSizeMillis += <-sizeChan
	}

	return Benchmark{
		url,
		int64(totalLatencyMillis / NUM_REQUESTS),
		int64(totalSizeMillis / NUM_REQUESTS),
	}, nil
}
