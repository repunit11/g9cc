package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type tokenKind int

const (
	tkReserved tokenKind = iota
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

func consume(op string, token *token) (bool, *token) {
	if token.kind != tkReserved || len(op) != token.len || token.str != op {
		return false, token
	}
	token = token.next
	return true, token
}

func expect(op string, tok *token) (*token, error) {
	if tok.kind != tkReserved || len(op) != tok.len || tok.str != op {
		return tok, fmt.Errorf("expected %q", op)
	}
	return tok.next, nil
}

func expectNumber(token *token) (int, *token, error) {
	if token.kind != tkNum {
		return 0, token, fmt.Errorf("expected a number")
	}
	val := token.val
	token = token.next
	return val, token, nil
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
				cur = newToken(tkReserved, cur, s[i:i+2], 2)
				i += 2
				continue
			}
		}

		// 記号の時トークン化
		if s[i] == '+' || s[i] == '-' || s[i] == '*' || s[i] == '/' || s[i] == '(' || s[i] == ')' || s[i] == '<' || s[i] == '>' {
			cur = newToken(tkReserved, cur, string(s[i]), 1)
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

type parser struct {
	tok *token
}

type nodeKind int

const (
	ndAdd nodeKind = iota
	ndSub
	ndMul
	ndDiv
	ndEq
	ndNe
	ndLt
	ndLe
	ndNum
)

type node struct {
	kind nodeKind
	lhs  *node
	rhs  *node
	val  int
}

func newNode(kind nodeKind, lhs *node, rhs *node) *node {
	node := &node{kind: kind, lhs: lhs, rhs: rhs}
	return node
}

func newNodeNum(val int) *node {
	node := &node{kind: ndNum, val: val}
	return node
}

func (p *parser) expr() *node {
	return p.equality()
}

func (p *parser) equality() *node {
	node := p.relational()
	var ok bool

	for {
		ok, p.tok = consume("==", p.tok)
		if ok {
			node = newNode(ndEq, node, p.relational())
			continue
		}
		ok, p.tok = consume("!=", p.tok)
		if ok {
			node = newNode(ndNe, node, p.relational())
			continue
		}
		return node
	}
}

func (p *parser) relational() *node {
	node := p.add()
	var ok bool

	for {
		ok, p.tok = consume("<", p.tok)
		if ok {
			node = newNode(ndLt, node, p.add())
			continue
		}
		ok, p.tok = consume("<=", p.tok)
		if ok {
			node = newNode(ndLe, node, p.add())
			continue
		}
		ok, p.tok = consume(">", p.tok)
		if ok {
			node = newNode(ndLt, p.add(), node)
			continue
		}
		ok, p.tok = consume(">=", p.tok)
		if ok {
			node = newNode(ndLe, p.add(), node)
			continue
		}
		return node
	}
}

func (p *parser) add() *node {
	node := p.mul()
	var ok bool

	for {
		ok, p.tok = consume("+", p.tok)
		if ok {
			node = newNode(ndAdd, node, p.mul())
			continue
		}
		ok, p.tok = consume("-", p.tok)
		if ok {
			node = newNode(ndSub, node, p.mul())
			continue
		}
		return node
	}
}

func (p *parser) mul() *node {
	node := p.unary()
	var ok bool

	for {
		ok, p.tok = consume("*", p.tok)
		if ok {
			node = newNode(ndMul, node, p.unary())
			continue
		}
		ok, p.tok = consume("/", p.tok)
		if ok {
			node = newNode(ndDiv, node, p.unary())
			continue
		}
		return node
	}
}

func (p *parser) unary() *node {
	var ok bool

	ok, p.tok = consume("+", p.tok)
	if ok {
		return p.primary()
	}

	ok, p.tok = consume("-", p.tok)
	if ok {
		return newNode(ndSub, newNodeNum(0), p.primary())
	}
	return p.primary()
}

func (p *parser) primary() *node {
	var err error
	var num int
	var ok bool

	ok, p.tok = consume("(", p.tok)
	if ok {
		node := p.expr()
		p.tok, err = expect(")", p.tok)
		if err != nil {
			return nil
		}
		return node
	}
	num, p.tok, err = expectNumber(p.tok)
	if err != nil {
		return nil
	}
	return newNodeNum(num)
}

func gen(node *node) {
	if node.kind == ndNum {
		fmt.Printf("	push %d\n", node.val)
		return
	}

	gen(node.lhs)
	gen(node.rhs)

	fmt.Printf("	pop rdi\n")
	fmt.Printf("	pop rax\n")

	switch node.kind {
	case ndAdd:
		fmt.Printf("	add rax, rdi\n")
		break
	case ndSub:
		fmt.Printf("	sub rax, rdi\n")
		break
	case ndMul:
		fmt.Printf("	imul rax, rdi\n")
		break
	case ndDiv:
		fmt.Printf("	cqo\n")
		fmt.Printf("	idiv rdi\n")
		break
	case ndEq:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	sete al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndNe:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setne al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndLt:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setl al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndLe:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setle al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	default:
		fmt.Fprintf(os.Stderr, "unexpected node kind")
		os.Exit(1)
	}
	fmt.Printf("	push rax\n")
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

	// パースする
	p := parser{tok: token}
	node := p.expr()

	// アセンブリの前半部分の出力
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// ASTの生成
	gen(node)

	fmt.Printf("	pop rax\n")
	fmt.Printf("	ret\n")
}
