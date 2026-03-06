package main

import "fmt"

func sema(prog *obj) error {
	for fn := prog; fn != nil; fn = fn.next {
		if err := addType(fn.body); err != nil {
			return err
		}
	}
	return nil
}

func intType() *ty {
	return &ty{kind: tyInt, size: 4}
}

func walk(nodes ...*node) error {
	for _, n := range nodes {
		if err := addType(n); err != nil {
			return err
		}
	}
	return nil
}

func normalizeArithmeticTypes(node *node) (lhsTy, rhsTy *ty) {
	lhsTy = node.lhs.ty
	rhsTy = node.rhs.ty

	if lhsTy.kind == tyArray {
		lhsTy = pointerTo(lhsTy.base)
	}
	if rhsTy.kind == tyArray {
		rhsTy = pointerTo(rhsTy.base)
	}
	return lhsTy, rhsTy
}

func scalePtrIndex(node *node, ptrTy *ty) error {
	scale := newNode(ndMul, node.rhs, newNodeNum(ptrTy.base.size))
	if err := addType(scale); err != nil {
		return err
	}
	node.rhs = scale
	node.ty = ptrTy
	return nil
}

func isIntegerType(t *ty) bool {
	return t.kind == tyInt || t.kind == tyChar
}

func typeAdd(node *node) error {
	lhsTy, rhsTy := normalizeArithmeticTypes(node)

	// num + num
	if isIntegerType(lhsTy) && isIntegerType(rhsTy) {
		node.ty = intType()
		return nil
	}

	// num + ptr to ptr + num
	if isIntegerType(lhsTy) && rhsTy.kind == tyPtr {
		node.lhs, node.rhs = node.rhs, node.lhs
		lhsTy, rhsTy = rhsTy, lhsTy
	}

	// ptr + num
	if lhsTy.kind == tyPtr && isIntegerType(rhsTy) {
		if err := scalePtrIndex(node, lhsTy); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("invalid operands for +")
}

func typeSub(node *node) error {
	lhsTy := node.lhs.ty
	rhsTy := node.rhs.ty

	if lhsTy.kind == tyArray {
		lhsTy = pointerTo(lhsTy.base)
	}
	if rhsTy.kind == tyArray {
		rhsTy = pointerTo(rhsTy.base)
	}

	// num - num
	if isIntegerType(lhsTy) && isIntegerType(rhsTy) {
		node.ty = intType()
		return nil
	}

	// ptr - num
	if lhsTy.kind == tyPtr && isIntegerType(rhsTy) {
		if err := scalePtrIndex(node, lhsTy); err != nil {
			return err
		}
		return nil
	}

	// ptr - ptr
	if lhsTy.kind == tyPtr && rhsTy.kind == tyPtr {
		sub := newNode(ndSub, node.lhs, node.rhs)
		sub.ty = intType()

		node.kind = ndDiv
		node.lhs = sub
		node.rhs = newNodeNum(lhsTy.base.size)
		node.ty = intType()
		return nil
	}

	return fmt.Errorf("invalid operands for -")
}

func addType(node *node) error {
	if node == nil {
		return nil
	}
	if err := walk(node.next, node.lhs, node.rhs, node.cond, node.then, node.els, node.init, node.inc); err != nil {
		return err
	}

	switch node.kind {
	case ndAdd:
		return typeAdd(node)
	case ndSub:
		return typeSub(node)
	case ndNe:
		node.ty = node.lhs.ty
		return nil
	case ndAssign:
		if node.lhs.ty.kind == tyArray {
			return fmt.Errorf("not an lvalue")
		}
		node.ty = node.lhs.ty
		return nil
	case ndMul, ndDiv, ndEq, ndLt, ndLe, ndNum:
		node.ty = intType()
		return nil
	case ndFuncall:
		for _, arg := range node.args {
			if err := addType(arg); err != nil {
				return err
			}
		}
		node.ty = intType()
		return nil
	case ndVar:
		node.ty = node.lvar.ty
		return nil
	case ndAddr:
		if node.lhs.ty.kind == tyArray {
			node.ty = pointerTo(node.lhs.ty.base)
		} else {
			node.ty = pointerTo(node.lhs.ty)
		}
		return nil
	case ndDeref:
		if node.lhs.ty.base != nil {
			node.ty = node.lhs.ty.base
		} else {
			node.ty = intType()
		}
		return nil
	case ndSizeof:
		node.ty = intType()
		node.kind = ndNum
		node.val = node.lhs.ty.size
		node.rhs = nil
		node.lhs = nil
	case ndExprStmt, ndReturn, ndIf, ndWhile, ndFor, ndBlock:
		return nil
	default:
		return fmt.Errorf("internal error: unknown node kind: %d", node.kind)
	}
	return nil
}
