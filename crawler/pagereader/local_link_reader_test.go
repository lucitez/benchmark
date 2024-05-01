package pagereader

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/lucitez/benchmark/helptest"
)

func TestLocalLinks(t *testing.T) {
	getPage := func(string) (*http.Response, error) {
		page := `<div><a href="/foo#what">foo</a><a href="/bar?huh">bar</a></div>`

		return &http.Response{Body: io.NopCloser(strings.NewReader(page))}, nil
	}

	pageReader, err := NewLocalLinkReader("https://go.dev", getPage)

	if err != nil {
		t.Fatal(err)
	}

	actual := pageReader.Read(func(s string) {})
	expected := []string{"https://go.dev/foo", "https://go.dev/bar"}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatal(helptest.Fail("LocalLinks()", fmt.Sprintf("%s", actual), fmt.Sprintf("%s", expected)))
	}
}
