package main

import (
	"fmt"
)

type parser struct {
	tok        *token
	locals     *LVar
	nextOffset int
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
	ndReturn
	ndIf
	ndWhile
	ndFor
	ndBlock
	ndFuncall
	ndAddr
	ndDeref
	ndSizeof
	ndNum
)

type node struct {
	kind     nodeKind // nodeの種類
	next     *node    // 次のnodeのアドレス
	lhs      *node    // 左子のnodeのアドレス
	rhs      *node    // 右子のnodeのアドレス
	lvar     *LVar    // ndVarの時に使用
	val      int      // ndNumの時に使用
	cond     *node    // if, forの時
	then     *node    // if, forの時
	els      *node    // ifの時
	init     *node    // forの時
	inc      *node    // forの時
	funcname string   // 関数名
	args     []*node  // 関数引数
	ty       *ty      // ポインタを表す型
}

type function struct {
	name   string
	params []*LVar
	body   *node
	next   *function
}

// LVar 連結リストで実装しているけど、マップの方が実装は楽そう
type LVar struct {
	next   *LVar
	name   string
	len    int
	offset int
	ty     *ty
}

func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

func stackAllocSize(t *ty) int {
	switch t.kind {
	case tyArray:
		return alignTo(t.arrayLen*8, 8)
	default:
		return 8
	}
}

func getNum(tok *token) (int, error) {
	if tok.kind != tkNum {
		return 0, fmt.Errorf("expected a number")
	}
	return tok.val, nil
}

func newNode(kind nodeKind, lhs *node, rhs *node) *node {
	node := &node{kind: kind, lhs: lhs, rhs: rhs}
	return node
}

func newNodeNum(val int) *node {
	node := &node{kind: ndNum, val: val}
	return node
}

func newFunc(name string, params []*LVar, body *node, next *function) *function {
	funct := &function{
		name:   name,
		params: params,
		body:   body,
		next:   next,
	}
	return funct
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

func (p *parser) findLVar(name string) *LVar {
	cur := p.locals
	for cur != nil {
		if cur.name == name {
			return cur
		}
		cur = cur.next
	}
	return nil
}

// funcdef = declspec ident "(" ( declspec ident ("," declspec ident)*)? ")" stmt
func (p *parser) funcdef() (*function, error) {
	if _, err := p.declspec(); err != nil {
		return nil, err
	}

	p.locals = nil
	p.nextOffset = 0
	if p.tok.kind == tkIdent {
		params := []*LVar{}
		funct := newFunc(p.tok.str, params, nil, nil)
		p.tok = p.tok.next
		if err := p.expect("("); err != nil {
			return nil, err
		}

		if p.tok.str != ")" {
			for {
				ty, err := p.declspec()
				if err != nil {
					return nil, err
				}
				if p.tok.kind != tkIdent {
					return nil, fmt.Errorf("expected ident")
				}
				tok := p.tok
				p.tok = p.tok.next

				lvar := p.findLVar(tok.str)
				if lvar != nil {
					return nil, fmt.Errorf("%s is already defined", tok.str)
				}
				p.nextOffset += stackAllocSize(ty)
				lvar = &LVar{
					name:   tok.str,
					len:    tok.len,
					offset: p.nextOffset,
					next:   p.locals,
					ty:     ty,
				}
				p.locals = lvar
				if len(params) >= len(argregs) {
					return nil, fmt.Errorf("too many parameters: max %d", len(argregs))
				}
				params = append(params, lvar)

				if !p.consume(",") {
					break
				}
			}
			funct.params = params
		}

		if err := p.expect(")"); err != nil {
			return nil, err
		}

		if p.tok.str != "{" {
			return nil, fmt.Errorf("expected {")
		}
		body, err := p.stmt()
		if err != nil {
			return nil, err
		}
		funct.body = body
		return funct, nil
	}
	return nil, fmt.Errorf("unexpected token")
}

// stmt = exprStmt
//
//	| "if" "(" expr ")" stmt ("else" stmt)?
//	| "return" expr ";"
//	| "while" "(" expr ")" stmt
//	| "for" "(" expr? ";" expr? ";" expr? ")" stmt
//	| "{" stmt* "}"
//	| ident "(" (ident ",")? ")" "{" stmt "}"
func (p *parser) stmt() (*node, error) {
	switch p.tok.kind {
	case tkIf:
		node := newNode(ndIf, nil, nil)
		p.tok = p.tok.next

		if err := p.expect("("); err != nil {
			return nil, err
		}
		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		node.cond = cond
		if err := p.expect(")"); err != nil {
			return nil, err
		}

		then, err := p.stmt()
		if err != nil {
			return nil, err
		}
		node.then = then

		if p.tok.kind == tkElse {
			p.tok = p.tok.next
			els, err := p.stmt()
			if err != nil {
				return nil, err
			}
			node.els = els
		}
		return node, nil
	case tkWhile:
		p.tok = p.tok.next

		if err := p.expect("("); err != nil {
			return nil, err
		}
		lhs, err := p.expr()
		if err != nil {
			return nil, err
		}

		if err := p.expect(")"); err != nil {
			return nil, err
		}
		rhs, err := p.stmt()
		if err != nil {
			return nil, err
		}
		node := newNode(ndWhile, lhs, rhs)
		return node, nil
	case tkReturn:
		p.tok = p.tok.next
		node, err := p.expr()
		if err != nil {
			return nil, err
		}
		node = newNode(ndReturn, node, nil)

		if err := p.expect(";"); err != nil {
			return nil, err
		}
		return node, nil
	case tkFor:
		p.tok = p.tok.next
		node := newNode(ndFor, nil, nil)
		if err := p.expect("("); err != nil {
			return nil, err
		}

		if p.tok.str != ";" {
			init, err := p.expr()
			if err != nil {
				return nil, err
			}
			node.init = init
		}
		if err := p.expect(";"); err != nil {
			return nil, err
		}

		if p.tok.str != ";" {
			cond, err := p.expr()
			if err != nil {
				return nil, err
			}
			node.cond = cond
		}
		if err := p.expect(";"); err != nil {
			return nil, err
		}

		if p.tok.str != ")" {
			inc, err := p.expr()
			if err != nil {
				return nil, err
			}
			node.inc = inc
		}
		if err := p.expect(")"); err != nil {
			return nil, err
		}

		then, err := p.stmt()
		if err != nil {
			return nil, err
		}
		node.then = then
		return node, nil
	case tkPunct:
		if p.consume("{") {
			head := new(node)
			cur := head
			for p.tok.str != "}" {
				next, err := p.stmt()
				if err != nil {
					return nil, err
				}
				cur.next = next
				cur = cur.next
			}
			if err := p.expect("}"); err != nil {
				return nil, err
			}
			node := newNode(ndBlock, head.next, nil)
			return node, nil
		}
	case tkInt:
		return p.declaration()
	}
	return p.exprStmt()
}

// declspec = "int"
func (p *parser) declspec() (*ty, error) {
	if p.tok.kind != tkInt {
		return nil, fmt.Errorf("expected type specifier 'int'")
	}
	p.tok = p.tok.next
	return &ty{kind: tyInt, size: 4}, nil
}

// type-suffix = "[" num "]" | ε
func (p *parser) typeSuffix(ty *ty) (*ty, error) {
	if p.consume("[") {
		sz, err := getNum(p.tok)
		if err != nil {
			return nil, err
		}
		p.tok = p.tok.next
		if err := p.expect("]"); err != nil {
			return nil, err
		}

		return arrayOf(ty, sz), nil
	}

	return ty, nil
}

// declarator = "*"* ident type-suffix
func (p *parser) declarator(ty *ty) (*ty, *token, error) {
	var err error
	for p.consume("*") {
		ty = pointerTo(ty)
	}

	if p.tok.kind != tkIdent {
		return nil, nil, fmt.Errorf("expected a variable name")
	}

	tok := p.tok
	p.tok = p.tok.next

	ty, err = p.typeSuffix(ty)
	if err != nil {
		return nil, nil, err
	}
	ty.name = p.tok

	return ty, tok, nil
}

// declaration = declspec declarator ";"
func (p *parser) declaration() (*node, error) {
	ty, err := p.declspec()
	if err != nil {
		return nil, err
	}

	ty, tok, err := p.declarator(ty)
	if err != nil {
		return nil, err
	}

	lvar := p.findLVar(tok.str)
	if lvar != nil {
		return nil, fmt.Errorf("%s is already defined", tok.str)
	}
	p.nextOffset += stackAllocSize(ty)
	lvar = &LVar{
		next:   p.locals,
		name:   tok.str,
		len:    tok.len,
		offset: p.nextOffset,
		ty:     ty,
	}
	p.locals = lvar

	if err := p.expect(";"); err != nil {
		return nil, err
	}

	return newNode(ndBlock, nil, nil), nil
}

// exprStmt = expr ";"
func (p *parser) exprStmt() (*node, error) {
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

// expr = assign
func (p *parser) expr() (*node, error) {
	return p.assign()
}

// assign = equality ("=" assign)*
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

// equality = relational ("==" relational | "!=" relational)*
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

// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
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

// add = mul ("+" mul | "-" mul)*
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

// mul = unary ("*" unary | "/" unary)*
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

// unary = ("+" | "-")? unary() | "*" unary | "&" unary | "sizeof" unary()
func (p *parser) unary() (*node, error) {
	if p.consume("+") {
		return p.unary()
	}

	if p.consume("-") {
		prim, err := p.unary()
		if err != nil {
			return nil, err
		}
		return newNode(ndSub, newNodeNum(0), prim), nil
	}

	if p.consume("*") {
		node, err := p.unary()
		if err != nil {
			return nil, err
		}
		node = newNode(ndDeref, node, nil)
		return node, nil
	}

	if p.consume("&") {
		node, err := p.unary()
		if err != nil {
			return nil, err
		}
		node = newNode(ndAddr, node, nil)
		return node, nil
	}

	if p.tok.kind == tkSizeof {
		p.tok = p.tok.next
		lhs, err := p.unary()
		if err != nil {
			return nil, err
		}
		node := newNode(ndSizeof, lhs, nil)
		return node, nil
	}

	return p.primary()
}

// primary = "(" expr ")" | number | ident ("(" (assign ("," assign)*)? ")")?
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
		name := p.tok.str
		p.tok = p.tok.next
		if p.consume("(") {
			node := newNode(ndFuncall, nil, nil)
			if !p.consume(")") {
				arg, err := p.assign()
				if err != nil {
					return nil, err
				}
				node.args = append(node.args, arg)
				for p.consume(",") {
					arg, err := p.assign()
					if err != nil {
						return nil, err
					}
					node.args = append(node.args, arg)
				}
				if err := p.expect(")"); err != nil {
					return nil, err
				}
			}

			node.funcname = name
			return node, nil
		}
		lvar := p.findLVar(name)
		if lvar == nil {
			return nil, fmt.Errorf("undefined variable: %s", name)
		}
		node := newNode(ndVar, nil, nil)
		node.lvar = lvar
		return node, nil
	}
	num, err := p.expectNumber()
	if err != nil {
		return nil, err
	}
	return newNodeNum(num), nil
}

func (p *parser) parse() (*function, error) {
	head := new(function)
	cur := head

	for p.tok.kind != tkEOF {
		n, err := p.funcdef()
		if err != nil {
			return nil, err
		}
		cur.next = n
		cur = cur.next
	}
	return head.next, nil
}
