package pagereader

import (
	"golang.org/x/net/html"
)

type PageReader interface {
	Tokenizer() (tokenizer *html.Tokenizer, close func())
	Read(func(string)) []string
}
