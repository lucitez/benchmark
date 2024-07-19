package client

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

// parse the host from the url
var HostSuffixRE = regexp.MustCompile(`([a-zA-Z0-9()]{1,256}\.)*([a-zA-Z0-9()]{1,256}\.[a-zA-Z0-9()]{1,6})`)

var Http = http.Client{
	Timeout: time.Second * 10,
	// do not allow redirects to a different host from the original request
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		fromSuffix := HostSuffixRE.FindSubmatch([]byte(via[0].URL.Host))
		toSuffix := HostSuffixRE.FindSubmatch([]byte(req.URL.Host))

		if len(fromSuffix) < 3 || len(toSuffix) < 3 {
			return fmt.Errorf("could not parse redirect from %s to %s", via[0].URL.Host, req.URL.Host)
		}

		if string(toSuffix[2]) != string(fromSuffix[2]) {
			return fmt.Errorf("skipping redirect from %s to %s", via[0].URL.Host, req.URL.Host)
		}

		if len(via) > 10 {
			return errors.New("to many redirects")
		}

		return nil
	},
}
