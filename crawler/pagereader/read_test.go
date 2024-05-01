package pagereader

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/lucitez/benchmark/helptest"
	"golang.org/x/net/html"
)

func TestSanitizeLocalHref(t *testing.T) {
	urls := []string{"/learn#anchor", "/learn?query"}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			actual := sanitizeLocalHref(url)
			expected := "/learn"

			if actual != expected {
				t.Fatal(helptest.Fail(fmt.Sprintf("sanitizeLocalHref(%s)", url), actual, expected))
			}
		})
	}
}

func TestGetHref(t *testing.T) {
	type test struct {
		a    string
		href string
	}

	tests := []test{{`<a href="foo">`, "foo"}, {`<a>`, ""}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("with tag %s", test.a), func(t *testing.T) {
			tzr := html.NewTokenizer(strings.NewReader(test.a))
			tzr.Next()

			actual := getHref(tzr)

			if actual != test.href {
				t.Fatal(helptest.Fail(fmt.Sprintf("getHref(%s)", test.a), actual, test.href))
			}
		})
	}

}

type MockParser struct{}

func (mp *MockParser) Tokenizer() (*html.Tokenizer, func()) {
	page := `<div><a href="foo">foo</a><a href="bar">bar</a></div>`

	tokenizer := html.NewTokenizer(strings.NewReader(page))

	return tokenizer, func() {}
}

func (mp *MockParser) URL() string {
	return "mock url"
}

func TestGetHrefs(t *testing.T) {
	mockParser := &MockParser{}

	actual := []string{}

	getHrefs(mockParser, func(s string) {
		actual = append(actual, s)
	})

	expected := []string{"foo", "bar"}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatal(helptest.Fail("getHref()", fmt.Sprintf("%s", actual), fmt.Sprintf("%s", expected)))
	}
}
