package main

import (
	"fmt"
	"strings"
)

func errorAt(input string, pos int, msg string) error {
	if pos < 0 {
		pos = 0
	}
	if pos > len(input) {
		pos = len(input)
	}
	caret := strings.Repeat(" ", pos) + "^"
	return fmt.Errorf("%s\n%s %s", input, caret, msg)
}
