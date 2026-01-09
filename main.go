package main

import (
	"fmt"
	"os"
	"strconv"
)

func readNumber(s string, i int) (int, int, error) {
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if start == i {
		return 0, i, fmt.Errorf("expected digit at %d", i)
	}
	n, err := strconv.Atoi(s[start:i])
	return n, i, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: g9cc <integer>")
		os.Exit(1)
	}
	rArg := os.Args[1]

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	num, next, err := readNumber(rArg, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("	mov rax, %d\n", num)
	i := next
	for i < len(rArg) {
		if rArg[i] == '+' {
			num, next, err := readNumber(rArg, i+1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			i = next
			fmt.Printf("	add rax, %d\n", num)
			continue
		} else if rArg[i] == '-' {
			num, next, err := readNumber(rArg, i+1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			i = next
			fmt.Printf("	sub rax, %d\n", num)
			continue
		}

		fmt.Fprintf(os.Stderr, "unexpected character: %q\n", rArg[i])
		os.Exit(1)
	}
	fmt.Printf("	ret\n")
}
