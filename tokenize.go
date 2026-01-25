package main

import (
	"fmt"
	"strconv"
)

type tokenKind int

const (
	tkPunct tokenKind = iota
	tkReturn
	tkIf
	tkElse
	tkWhile
	tkFor
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

var doublePunct = map[string]struct{}{
	"==": {},
	"!=": {},
	"<=": {},
	">=": {},
}

var singlePunct = map[byte]struct{}{
	'+': {},
	'-': {},
	'*': {},
	'/': {},
	'(': {},
	')': {},
	'<': {},
	'>': {},
	';': {},
	'=': {},
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

func isSinglePunct(b byte) bool {
	_, ok := singlePunct[b]
	return ok
}

func isDoublePunct(s string, i int) (string, bool) {
	if i+1 >= len(s) {
		return "", false
	}
	tok := s[i : i+2]
	_, ok := doublePunct[tok]
	return tok, ok
}

func isIdentStart(b byte) bool {
	return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || b == '_'
}

func isIdentCont(b byte) bool {
	return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || b == '_' || ('0' <= b && b <= '9')
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

		// 識別子・予約語の時トークン化
		if isIdentStart(s[i]) {
			j := i
			j++
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			ident := s[i:j]
			kind := tkIdent

			switch ident {
			case "return":
				kind = tkReturn
			case "if":
				kind = tkIf
			case "else":
				kind = tkElse
			case "while":
				kind = tkWhile
			case "for":
				kind = tkFor
			}
			cur = newToken(kind, cur, ident, len(ident))
			i += len(ident)
			continue
		}

		// 複数文字のトークン化
		if tok, ok := isDoublePunct(s, i); ok {
			cur = newToken(tkPunct, cur, tok, 2)
			i += 2
			continue
		}

		// 記号の時トークン化
		if isSinglePunct(s[i]) {
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

		return nil, errorAt(s, i, "unexpected token")
	}
	// 末尾文字をつけてトークン化
	newToken(tkEOF, cur, "", 0)
	return head.next, nil
}
