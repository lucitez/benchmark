package helptest

import (
	"fmt"
)

func Fail(call, actual, expected string) string {
	return fmt.Sprintf("\nTest of %s:\nActual: %s\nExpctd: %s\n", call, actual, expected)
}
