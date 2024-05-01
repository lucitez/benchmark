package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lucitez/benchmark/crawler/pagereader"
)

var MAX_DEPTH = 2
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
Crawl recursively visits urls until it reaches the MAX_DEPTH or runs out of urls to crawl.

url: the url to visit. caller should pass the root url of the website to crawl.
depth: the depth of the current search. caller *must* pass 0.
out: the channel to which each visited url will be sent
safemap: a way to cache visited urls. caller can pass an empty sync.Map pointer

!!! depth _must_ be passed as 0 by the original caller, or else the channel will never close. !!!
*/
func Crawl(url string, depth int, out chan<- string, safemap *sync.Map) {
	if depth == 0 {
		defer close(out)
	}

	if depth >= MAX_DEPTH {
		return
	}

	llr, err := pagereader.NewLocalLinkReader(url, client.Get)

	// There was a problem accessing the url, likely due to a disallowed redirect
	if err != nil {
		// fmt.Println(err) this can get noisy. TODO add a verbose flag
		return
	}

	out <- url

	wg := sync.WaitGroup{}
	llr.Read(func(foundUrl string) {
		// do not visit urls with file extensions. TODO allow .html files?
		var re = regexp.MustCompile(`.*\.\w{2,}$`)

		if matches := re.Find([]byte(foundUrl)); matches != nil {
			return
		}

		if _, loaded := safemap.LoadOrStore(foundUrl, true); loaded {
			return
		}

		wg.Add(1)

		go func() {
			Crawl(foundUrl, depth+1, out, safemap)
			wg.Done()
		}()
	})

	wg.Wait()
}
