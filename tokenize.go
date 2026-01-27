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
	'{': {},
	'}': {},
	',': {},
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

func scanIdentOrKeyword(s string, i int) (*token, int, bool) {
	if !isIdentStart(s[i]) {
		return nil, i, false
	}
	j := i + 1
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
	return newToken(kind, ident, len(ident)), j, true
}

func scanDoublePunct(s string, i int) (*token, int, bool) {
	val, ok := isDoublePunct(s, i)
	if !ok {
		return nil, i, ok
	}
	j := i + 2
	return newToken(tkPunct, val, len(val)), j, ok
}

func scanSinglePunct(s string, i int) (*token, int, bool) {
	ok := isSinglePunct(s[i])
	if !ok {
		return nil, i, ok
	}
	j := i + 1
	return newToken(tkPunct, string(s[i]), 1), j, ok
}

func scanNumber(s string, i int) (*token, int, bool, error) {
	if !isDigit(s[i]) {
		return nil, i, false, nil
	}
	num, next, err := readNumber(s, i)
	if err != nil {
		return nil, i, false, err
	}
	val, err := strconv.Atoi(num)
	if err != nil {
		return nil, i, false, err
	}
	tok := newToken(tkNum, num, 1)
	tok.val = val
	return tok, next, true, nil
}

func newToken(kind tokenKind, str string, len int) *token {
	return &token{kind: kind, str: str, len: len}
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
		if tok, next, ok := scanIdentOrKeyword(s, i); ok {
			cur.next = tok
			cur = tok
			i = next
			continue
		}

		// 複数文字のトークン化
		if tok, next, ok := scanDoublePunct(s, i); ok {
			cur.next = tok
			cur = tok
			i = next
			continue
		}

		// 記号の時トークン化
		if tok, next, ok := scanSinglePunct(s, i); ok {
			cur.next = tok
			cur = tok
			i = next
			continue
		}

		// 数字の時トークン化
		if tok, next, ok, err := scanNumber(s, i); err != nil {
			return nil, err
		} else if ok {
			cur.next = tok
			cur = tok
			i = next
			continue
		}

		return nil, errorAt(s, i, "unexpected token")
	}
	// 末尾文字をつけてトークン化
	cur.next = newToken(tkEOF, "", 0)
	return head.next, nil
}
