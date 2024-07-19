package client

import (
	"fmt"
	"net/url"
	"testing"
)

func TestRegexMatching(t *testing.T) {
	testStr := "https://go.dev/pkg/route"
	testStr2 := "https://pkg.go.dev/some/route"

	u, _ := url.ParseRequestURI(testStr)
	fmt.Println(u.Host)

	u, _ = url.ParseRequestURI(testStr2)
	fmt.Println(u.Host)

	submatch := HostSuffixRE.FindSubmatch([]byte(testStr))
	fmt.Printf("For %s\n", testStr)
	for _, s := range submatch {
		fmt.Printf("Submatch: %s\n", s)
	}

	submatch = HostSuffixRE.FindSubmatch([]byte(testStr2))
	fmt.Printf("For %s\n", testStr)
	for _, s := range submatch {
		fmt.Printf("Submatch: %s\n", s)
	}
}
