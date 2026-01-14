package main

import (
	"fmt"
	"os"
	"strconv"
)

type tokenKind int

const (
	TK_RESERVED tokenKind = iota
	TK_NUM
	TK_EOF
)

type token struct {
	kind tokenKind
	next *token
	val  int
	str  string
}

func readNumber(s string, i int) (string, int, error) {
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if start == i {
		return "0", i, fmt.Errorf("expected digit at %d", i)
	}
	return s[start:i], i, nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func consume(op uint8, token *token) (bool, *token) {
	if token.kind != TK_RESERVED || token.str[0] != op {
		return false, token
	}
	token = token.next
	return true, token
}

func expectNumber(token *token) (int, *token, error) {
	if token.kind != TK_NUM {
		return 0, token, fmt.Errorf("not number")
	}
	val := token.val
	token = token.next
	return val, token, nil
}

func newToken(kind tokenKind, cur *token, str string) *token {
	tok := &token{kind: kind, str: str}
	cur.next = tok
	return tok
}

func tokenize(s string) (*token, error) {
	head := token{next: nil}
	cur := &head
	i := 0

	for i < len(s) {
		// 記号の時トークン化
		if s[i] == '+' || s[i] == '-' {
			cur = newToken(TK_RESERVED, cur, string(s[i]))
			i++
			continue
		}

		// 数字の時トークン化
		if isDigit(byte(s[i])) {
			num, next, err := readNumber(s, i)
			if err != nil {
				return nil, err
			}
			cur = newToken(TK_NUM, cur, num)
			val, err := strconv.Atoi(num)
			if err != nil {
				return nil, err
			}
			cur.val = val
			i = next
			continue
		}
		return nil, fmt.Errorf("unexpected character: %q\n", s[i])
	}
	// 末尾文字をつけてトークン化
	newToken(TK_EOF, cur, "")
	return head.next, nil
}

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

	// アセンブリの前半部分の出力
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	num, token, err := expectNumber(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("	mov rax, %d\n", num)

	// トークンを消費してアセンブリを出力
	for token.kind != TK_EOF {
		var ok bool
		ok, token = consume('+', token)
		if ok {
			num, token, err = expectNumber(token)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("	add rax, %d\n", num)
			continue
		}
		ok, token = consume('-', token)
		if ok {
			num, token, err = expectNumber(token)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("	sub rax, %d\n", num)
			continue
		}
		fmt.Fprintln(os.Stderr, "unexpected token")
		os.Exit(1)
	}
	fmt.Printf("	ret\n")
}
