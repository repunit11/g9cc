package main

import "fmt"

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
	ndAssign
	ndExprStmt
	ndVar
	ndNum
)

type node struct {
	kind nodeKind // nodeの種類
	next *node    // 次のnodeのアドレス
	lhs  *node    // 左子のnodeのアドレス
	rhs  *node    // 右子のnodeのアドレス
	name byte     // ndVarの時に使用
	val  int      // ndNumの時に使用
}

func newNode(kind nodeKind, lhs *node, rhs *node) *node {
	node := &node{kind: kind, lhs: lhs, rhs: rhs}
	return node
}

func newNodeNum(val int) *node {
	node := &node{kind: ndNum, val: val}
	return node
}

func (p *parser) consume(op string) bool {
	if p.tok.kind != tkPunct || len(op) != p.tok.len || p.tok.str != op {
		return false
	}
	p.tok = p.tok.next
	return true
}

func (p *parser) expect(op string) error {
	if p.tok.kind != tkPunct || len(op) != p.tok.len || p.tok.str != op {
		return fmt.Errorf("expected %q", op)
	}
	p.tok = p.tok.next
	return nil
}

func (p *parser) expectNumber() (int, error) {
	if p.tok.kind != tkNum {
		return 0, fmt.Errorf("expected a number")
	}
	val := p.tok.val
	p.tok = p.tok.next
	return val, nil
}

func (p *parser) stmt() (*node, error) {
	return p.expr_stmt()
}

func (p *parser) expr_stmt() (*node, error) {
	node, err := p.expr()
	if err != nil {
		return nil, err
	}

	node = newNode(ndExprStmt, node, nil)

	if err := p.expect(";"); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *parser) expr() (*node, error) {
	return p.assign()
}

func (p *parser) assign() (*node, error) {
	node, err := p.equality()
	if err != nil {
		return nil, err
	}
	for {
		if p.consume("=") {
			rhs, err := p.assign()
			if err != nil {
				return nil, err
			}
			node = newNode(ndAssign, node, rhs)
			continue
		}
		return node, nil
	}
}

func (p *parser) equality() (*node, error) {
	node, err := p.relational()
	if err != nil {
		return nil, err
	}

	for {
		if p.consume("==") {
			rhs, err := p.relational()
			if err != nil {
				return nil, err
			}
			node = newNode(ndEq, node, rhs)
			continue
		}
		if p.consume("!=") {
			rhs, err := p.relational()
			if err != nil {
				return nil, err
			}
			node = newNode(ndNe, node, rhs)
			continue
		}
		return node, nil
	}
}

func (p *parser) relational() (*node, error) {
	node, err := p.add()
	if err != nil {
		return nil, err
	}

	for {
		if p.consume("<") {
			rhs, err := p.add()
			if err != nil {
				return nil, err
			}
			node = newNode(ndLt, node, rhs)
			continue
		}
		if p.consume("<=") {
			rhs, err := p.add()
			if err != nil {
				return nil, err
			}
			node = newNode(ndLe, node, rhs)
			continue
		}
		if p.consume(">") {
			lhs, err := p.add()
			if err != nil {
				return nil, err
			}
			node = newNode(ndLt, lhs, node)
			continue
		}
		if p.consume(">=") {
			lhs, err := p.add()
			if err != nil {
				return nil, err
			}
			node = newNode(ndLe, lhs, node)
			continue
		}
		return node, nil
	}
}

func (p *parser) add() (*node, error) {
	node, err := p.mul()
	if err != nil {
		return nil, err
	}

	for {
		if p.consume("+") {
			rhs, err := p.mul()
			if err != nil {
				return nil, err
			}
			node = newNode(ndAdd, node, rhs)
			continue
		}
		if p.consume("-") {
			rhs, err := p.mul()
			if err != nil {
				return nil, err
			}
			node = newNode(ndSub, node, rhs)
			continue
		}
		return node, nil
	}
}

func (p *parser) mul() (*node, error) {
	node, err := p.unary()
	if err != nil {
		return nil, err
	}

	for {
		if p.consume("*") {
			rhs, err := p.unary()
			if err != nil {
				return nil, err
			}
			node = newNode(ndMul, node, rhs)
			continue
		}
		if p.consume("/") {
			rhs, err := p.unary()
			if err != nil {
				return nil, err
			}
			node = newNode(ndDiv, node, rhs)
			continue
		}
		return node, nil
	}
}

func (p *parser) unary() (*node, error) {
	if p.consume("+") {
		return p.primary()
	}

	if p.consume("-") {
		prim, err := p.primary()
		if err != nil {
			return nil, err
		}
		return newNode(ndSub, newNodeNum(0), prim), nil
	}
	return p.primary()
}

func (p *parser) primary() (*node, error) {
	if p.consume("(") {
		node, err := p.expr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(")"); err != nil {
			return nil, err
		}
		return node, nil
	}

	if p.tok.kind == tkIdent {
		node := newNode(ndVar, nil, nil)
		node.name = p.tok.str[0]
		p.tok = p.tok.next
		return node, nil
	}
	num, err := p.expectNumber()
	if err != nil {
		return nil, err
	}
	return newNodeNum(num), nil
}

func (p *parser) parse() (*node, error) {
	head := new(node)
	cur := head

	for p.tok.kind != tkEOF {
		n, err := p.stmt()
		if err != nil {
			return nil, err
		}
		cur.next = n
		cur = cur.next
	}
	return head.next, nil
}
