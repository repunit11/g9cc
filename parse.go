package main

import (
	"fmt"
)

type parser struct {
	tok        *token
	locals     *obj
	nextOffset int
	input      string
	globals    *obj
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
	lvar     *obj     // ndVarの時に使用
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

type obj struct {
	next       *obj
	name       *string // Variable name
	ty         *ty     // type
	isLocal    bool    // local or global
	offset     int     // local variable
	isFunction bool    // global variable or function
	// function
	params    *obj
	body      *node
	locals    *obj
	stackSize int
}

func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

func stackAllocSize(t *ty) int {
	switch t.kind {
	case tyArray:
		return alignTo(t.size, 8)
	default:
		return t.size
	}
}

func getNum(tok *token) (int, error) {
	if tok.kind != tkNum {
		return 0, fmt.Errorf("expected a number")
	}
	return tok.val, nil
}

func (p *parser) declareLocal(tok *token, ty *ty) (*obj, error) {
	lvar := p.findLVar(tok.str)
	if lvar != nil {
		return nil, errorAt(p.input, tok.pos, fmt.Sprintf("%s is already defined", tok.str))
	}
	p.nextOffset += stackAllocSize(ty)
	lvar = &obj{
		next:    p.locals,
		name:    &tok.str,
		offset:  p.nextOffset,
		ty:      ty,
		isLocal: true,
	}
	p.locals = lvar
	return lvar, nil
}

func newNode(kind nodeKind, lhs *node, rhs *node) *node {
	node := &node{kind: kind, lhs: lhs, rhs: rhs}
	return node
}

func newNodeNum(val int) *node {
	node := &node{kind: ndNum, val: val}
	return node
}

func newFunc(name string, params *obj, body *node, next *obj) *obj {
	funct := &obj{
		name:       &name,
		params:     params,
		body:       body,
		next:       next,
		isFunction: true,
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
		return errorAt(p.input, p.tok.pos, fmt.Sprintf("expected %q", op))
	}
	p.tok = p.tok.next
	return nil
}

func (p *parser) expectNumber() (int, error) {
	if p.tok.kind != tkNum {
		return 0, errorAt(p.input, p.tok.pos, "expected a number")
	}
	val := p.tok.val
	p.tok = p.tok.next
	return val, nil
}

func (p *parser) findLVar(name string) *obj {
	cur := p.locals
	for cur != nil {
		if *cur.name == name {
			return cur
		}
		cur = cur.next
	}
	return nil
}

func (p *parser) findGVar(name string) *obj {
	cur := p.globals
	for cur != nil {
		if *cur.name == name {
			return cur
		}
		cur = cur.next
	}
	return nil
}

func newVar(name string, ty *ty) *obj {
	return &obj{name: &name, ty: ty}
}

func (p *parser) newGVar(name string, ty *ty) *obj {
	v := newVar(name, ty)
	v.isLocal = false
	v.next = p.globals
	p.globals = v
	return v
}

func reverseObjList(head *obj) *obj {
	var prev *obj
	cur := head
	for cur != nil {
		next := cur.next
		cur.next = prev
		prev = cur
		cur = next
	}
	return prev
}

// funcdef = declspec ident "(" ( declspec ident ("," declspec ident)*)? ")" stmt
func (p *parser) funcdef() (*obj, error) {
	if _, err := p.declspec(); err != nil {
		return nil, err
	}

	p.locals = nil
	p.nextOffset = 0
	if p.tok.kind == tkIdent {
		funct := newFunc(p.tok.str, nil, nil, nil)
		p.tok = p.tok.next
		if err := p.expect("("); err != nil {
			return nil, err
		}

		if p.tok.str != ")" {
			nparams := 0
			for {
				ty, err := p.declspec()
				if err != nil {
					return nil, err
				}
				if p.tok.kind != tkIdent {
					return nil, errorAt(p.input, p.tok.pos, "expected identifier")
				}
				tok := p.tok
				p.tok = p.tok.next

				_, err = p.declareLocal(tok, ty)
				if err != nil {
					return nil, err
				}

				if nparams >= len(argregs64) {
					return nil, errorAt(p.input, p.tok.pos, fmt.Sprintf("too many parameters: max %d", len(argregs64)))
				}
				nparams++

				if !p.consume(",") {
					break
				}
			}
			funct.params = reverseObjList(p.locals)
			p.locals = funct.params
		}

		if err := p.expect(")"); err != nil {
			return nil, err
		}

		if p.tok.str != "{" {
			return nil, errorAt(p.input, p.tok.pos, "expected {")
		}
		body, err := p.stmt()
		if err != nil {
			return nil, err
		}
		funct.body = body
		return funct, nil
	}
	return nil, errorAt(p.input, p.tok.pos, "unexpected token")
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
		return nil, errorAt(p.input, p.tok.pos, "expected type specifier 'int'")
	}
	p.tok = p.tok.next
	return &ty{kind: tyInt, size: 4}, nil
}

// type-suffix = "(" ( declspec ident ("," declspec ident)*)? ")" | "[" num "]" type-suffix | ε
func (p *parser) typeSuffix(ty *ty) (*ty, error) {
	if p.consume("(") {
		if !p.consume(")") {
			for {
				if _, err := p.declspec(); err != nil {
					return nil, err
				}
				if p.tok.kind != tkIdent {
					return nil, errorAt(p.input, p.tok.pos, "expected identifier")
				}
				p.tok = p.tok.next

				if !p.consume(",") {
					break
				}
			}
			if err := p.expect(")"); err != nil {
				return nil, err
			}
		}
		return funcType(ty), nil
	}

	if p.consume("[") {
		sz, err := getNum(p.tok)
		if err != nil {
			return nil, err
		}
		p.tok = p.tok.next
		if err := p.expect("]"); err != nil {
			return nil, err
		}

		ty, err = p.typeSuffix(ty)
		if err != nil {
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
		return nil, nil, errorAt(p.input, p.tok.pos, "expected a variable name")
	}

	tok := p.tok
	p.tok = p.tok.next

	ty, err = p.typeSuffix(ty)
	if err != nil {
		return nil, nil, err
	}
	ty.name = tok

	return ty, tok, nil
}

// declaration = declspec (declarator ("=" expr)? ("," declarator ("=" expr)?)*)? ";"
func (p *parser) declaration() (*node, error) {
	basety, err := p.declspec()
	if err != nil {
		return nil, err
	}

	head := new(node)
	cur := head
	if p.consume(";") {
		return newNode(ndBlock, head.next, nil), nil
	}
	for {
		ty, tok, err := p.declarator(basety)
		if err != nil {
			return nil, err
		}

		lvar, err := p.declareLocal(tok, ty)
		if err != nil {
			return nil, err
		}

		if p.consume("=") {
			rhs, err := p.expr()
			if err != nil {
				return nil, err
			}

			lhs := newNode(ndVar, nil, nil)
			lhs.lvar = lvar

			assign := newNode(ndAssign, lhs, rhs)
			stmt := newNode(ndExprStmt, assign, nil)

			cur.next = stmt
			cur = cur.next
		}
		if !p.consume(",") {
			break
		}
	}

	if err := p.expect(";"); err != nil {
		return nil, err
	}
	return newNode(ndBlock, head.next, nil), nil
}

// exprStmt = expr? ";"
func (p *parser) exprStmt() (*node, error) {
	if p.consume(";") {
		return newNode(ndBlock, nil, nil), nil
	}

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

// unary = ("+" | "-")? unary() | "*" unary | "&" unary | "sizeof" unary | postfix
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

	return p.postfix()
}

// postfix = primary ("[" expr "]")*
func (p *parser) postfix() (*node, error) {
	node, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.consume("[") {
			rhs, err := p.expr()
			if err != nil {
				return nil, err
			}

			if err = p.expect("]"); err != nil {
				return nil, err
			}
			// x[y] => *(x + y)
			add := newNode(ndAdd, node, rhs)
			node = newNode(ndDeref, add, nil)
			continue
		}
		return node, nil
	}
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
		tok := p.tok
		name := tok.str
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
			lvar = p.findGVar(name)
			if lvar == nil {
				return nil, errorAt(p.input, tok.pos, fmt.Sprintf("undefined variable: %s", name))
			}
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

func (p *parser) globalVariable() error {
	basety, err := p.declspec()
	if err != nil {
		return err
	}
	for {
		ty, tok, err := p.declarator(basety)
		if err != nil {
			return err
		}
		p.newGVar(tok.str, ty)

		if !p.consume(",") {
			break
		}
	}
	return p.expect(";")
}

func isFunction(tok *token) bool {
	p := &parser{tok: tok}
	basety, err := p.declspec()
	if err != nil {
		return false
	}
	ty, _, err := p.declarator(basety)
	if err != nil {
		return false
	}
	return ty.kind == tyFunc

}

func (p *parser) parse() (*obj, error) {
	head := new(obj)
	cur := head

	for p.tok.kind != tkEOF {
		if isFunction(p.tok) {
			fn, err := p.funcdef()
			if err != nil {
				return nil, err
			}
			cur.next = fn
			cur = cur.next
			continue
		}

		if err := p.globalVariable(); err != nil {
			return nil, err
		}
	}

	if head.next == nil {
		return p.globals, nil
	}

	prog := head.next
	for cur.next != nil {
		cur = cur.next
	}
	cur.next = p.globals
	return prog, nil
}
