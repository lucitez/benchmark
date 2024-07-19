package crawler

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type pageReader struct {
	url  *url.URL
	page *http.Response
}

func newPageReader(rawUrl string, getFunc func(string) (*http.Response, error)) (pageReader, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return pageReader{}, err
	}

	resp, err := getFunc(parsedUrl.String())
	if err != nil {
		return pageReader{}, err
	}

	localLinkReader := pageReader{
		url:  parsedUrl,
		page: resp,
	}

	return localLinkReader, nil
}

// ScrapeLocalURLs finds all hrefs to the same host
// Todo make scrape functions that take a page and return different things?
func (r pageReader) ScrapeLocalURLs(out chan<- string) []string {
	defer r.page.Body.Close()

	urls := []string{}

	tokenizer := html.NewTokenizer(r.page.Body)

	if tokenizer == nil {
		fmt.Printf("Could not create tokenizer for %s\n", r.url.String())
		return urls
	}

	for {
		tt := tokenizer.Next()

		switch tt {
		case html.ErrorToken:
			close(out)
			return urls
		case html.StartTagToken:
			tag, hasAttr := tokenizer.TagName()

			if string(tag) == "a" && hasAttr {
				tags := getTags(tokenizer)

				if href, ok := tags["href"]; ok {
					out <- href
					urls = append(urls, href)
				}
			}
		}
	}
}

func getTags(tokenizer *html.Tokenizer) map[string]string {
	tags := map[string]string{}

	for attr, val, next := tokenizer.TagAttr(); true; attr, val, next = tokenizer.TagAttr() {
		tags[string(attr)] = string(val)

		if !next {
			break
		}
	}

	return tags
}
