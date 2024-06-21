package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lucitez/benchmark/pagereader"
)

type Crawler struct {
	rootURL  string
	maxDepth int
}

func New(rootURL string, maxDepth int) Crawler {
	return Crawler{
		rootURL,
		maxDepth,
	}
}

var client = http.Client{
	Timeout: time.Second * 5,
	// do not allow redirects to a different host from the original request
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if strings.TrimPrefix(req.URL.Host, "www.") != strings.TrimPrefix(via[0].URL.Host, "www.") {
			return fmt.Errorf("skipping redirect from %s to %s", via[0].URL.Host, req.URL.Host)
		}

		if len(via) > 10 {
			return errors.New("to many redirects")
		}

		return nil
	},
}

/*
Crawl will crawl the crawler's url and send any url visited to chan visited
*/
func (c Crawler) Crawl(visited chan<- string) {
	visitedMap := &sync.Map{}
	visitedMap.Store(c.rootURL, true)
	go c.crawl(c.rootURL, 0, visited, visitedMap)
}

/*
crawl recursively visits urls until it reaches the MAX_DEPTH or runs out of urls to crawl.

url: the url to visit. caller should pass the root url of the website to crawl.
depth: the depth of the current search. caller *must* pass 0.
out: the channel to which each visited url will be sent
safemap: a way to cache visited urls. caller can pass an empty sync.Map pointer

!!! depth _must_ be passed as 0 by the original caller, or else the channel will never close. !!!
*/
func (c Crawler) crawl(urlStr string, depth int, urlOut chan<- string, visited *sync.Map) {
	// this recursive function exits once all subroutines have finished
	if depth == 0 {
		defer close(urlOut)
	}

	if depth >= c.maxDepth {
		return
	}

	pageReader, err := pagereader.New(urlStr, client.Get)
	// TODO send these to an error chan
	if err != nil {
		fmt.Printf("Error getting page for %s: %v\n", urlStr, err)
		return
	}

	// At this point, consider the current url visited
	urlOut <- urlStr

	rootUrl, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("Error parsing url into struct %v\n", err)
	}

	foundUrlChan := make(chan string)
	go pageReader.ScrapeLocalURLs(foundUrlChan)

	wg := sync.WaitGroup{}
	for foundUrl := range foundUrlChan {
		sanitizedUrl, ok := validateUrl(rootUrl, foundUrl)
		if !ok {
			continue
		}

		// we have already visited this url, skip it
		if _, loaded := visited.LoadOrStore(sanitizedUrl, true); loaded {
			continue
		}

		wg.Add(1)
		go func(u string) {
			c.crawl(u, depth+1, urlOut, visited)
			wg.Done()
		}(sanitizedUrl)
	}

	wg.Wait()
}

func validateUrl(rootUrl *url.URL, foundUrl string) (string, bool) {
	// exclude urls with file extensions
	// TODO: allow .html files?
	var re = regexp.MustCompile(`.*\.\w{2,}$`)
	if matches := re.Find([]byte(foundUrl)); matches != nil {
		return "", false
	}

	fullFoundUrl, err := rootUrl.Parse(foundUrl)
	if err != nil {
		return "", false
	}

	// Only follow urls with same host
	if !strings.HasSuffix(fullFoundUrl.Host, rootUrl.Host) {
		return "", false
	}

	sanitizedUrl := sanitizeUrl(fullFoundUrl.String())

	return sanitizedUrl, sanitizedUrl != ""
}

// strip query strings and anchor tags
func sanitizeUrl(url string) string {
	var re = regexp.MustCompile(`([^#?]*)[#?]?.*`)

	matches := re.FindSubmatch([]byte(url))

	if len(matches) != 2 {
		fmt.Printf("Could not sanitize url: %s\n", url)
		return ""
	}

	match := string(matches[1])
	match = strings.TrimSuffix(match, "/")
	return match
}
