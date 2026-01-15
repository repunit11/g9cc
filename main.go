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
	p := parser{tok: token}
	node, err := p.expr()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// アセンブリの前半部分の出力
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// ASTの生成
	gen(node)

	fmt.Printf("	pop rax\n")
	fmt.Printf("	ret\n")
}
