package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: g9cc <integer>")
		os.Exit(1)
	}
	rArg := os.Args[1]

	// トークナイズする
	token, err := tokenize(rArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// パースする
	p := parser{tok: token, locals: nil, nextOffset: 0}
	functs, err := p.parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// アセンブリの生成
	codegen(functs)
}
