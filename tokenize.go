package main

import (
	"fmt"
	"strconv"
)

type tokenKind int

const (
	tkPunct tokenKind = iota
	tkIdent
	tkNum
	tkEOF
)

type token struct {
	kind tokenKind
	next *token
	val  int
	str  string
	len  int
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

func newToken(kind tokenKind, cur *token, str string, len int) *token {
	tok := &token{kind: kind, str: str, len: len}
	cur.next = tok
	return tok
}

func tokenize(s string) (*token, error) {
	head := token{next: nil}
	cur := &head
	i := 0

	for i < len(s) {
		// 空白の時スキップ
		if s[i] == ' ' {
			i++
			continue
		}

		// 複数文字のトークン化
		if i+1 < len(s) {
			switch s[i : i+2] {
			case "==", "!=", "<=", ">=":
				cur = newToken(tkPunct, cur, s[i:i+2], 2)
				i += 2
				continue
			}
		}

		// 記号の時トークン化
		if s[i] == '+' || s[i] == '-' || s[i] == '*' || s[i] == '/' || s[i] == '(' || s[i] == ')' || s[i] == '<' || s[i] == '>' || s[i] == ';' || s[i] == '=' {
			cur = newToken(tkPunct, cur, string(s[i]), 1)
			i++
			continue
		}

		// 数字の時トークン化
		if isDigit(s[i]) {
			num, next, err := readNumber(s, i)
			if err != nil {
				return nil, err
			}
			cur = newToken(tkNum, cur, num, 1)
			val, err := strconv.Atoi(num)
			if err != nil {
				return nil, err
			}
			cur.val = val
			i = next
			continue
		}

		// 1文字のローカル変数の時トークン化
		if 'a' <= s[i] && s[i] <= 'z' {
			cur = newToken(tkIdent, cur, string(s[i]), 1)
			i++
			continue
		}

		return nil, errorAt(s, i, "unexpected token")
	}
	// 末尾文字をつけてトークン化
	newToken(tkEOF, cur, "", 0)
	return head.next, nil
}
