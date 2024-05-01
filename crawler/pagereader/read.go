package pagereader

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// only take the path before query strings, anchors, etc.
// TODO make this better
func sanitizeLocalHref(href string) string {
	var re = regexp.MustCompile(`(/[-_\.\da-zA-Z]*)+`)

	matches := re.FindSubmatch([]byte(href))

	match := string(matches[0])
	match = strings.TrimSuffix(match, "/")
	return match
}

type Tokenizerer interface {
	URL() string
	Tokenizer() (tokenizer *html.Tokenizer, close func())
}

func getHrefs(tk Tokenizerer, processHref func(string)) {
	tokenizer, close := tk.Tokenizer()

	if tokenizer == nil {
		fmt.Printf("Could not create tokenizer for %s\n", tk.URL())
		return
	}

	for {
		tt := tokenizer.Next()

		switch tt {
		case html.ErrorToken:
			close()
			return
		case html.StartTagToken:
			tag, hasAttr := tokenizer.TagName()

			if string(tag) == "a" && hasAttr {
				href := getHref(tokenizer)
				processHref(href)
			}
		}
	}
}

func getHref(tokenizer *html.Tokenizer) string {
	for attr, val, next := tokenizer.TagAttr(); true; attr, val, next = tokenizer.TagAttr() {
		if string(attr) == "href" {
			return string(val)
		}

		if !next {
			break
		}
	}

	return ""
}
