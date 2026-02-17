package main

import "fmt"

func sema(prog *function) error {
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

func typeAdd(node *node) error {
	// num + num
	if node.lhs.ty.kind == tyInt && node.rhs.ty.kind == tyInt {
		node.ty = intType()
		return nil
	}

	// num + ptr to ptr + num
	if node.lhs.ty.kind == tyInt && node.rhs.ty.kind == tyPtr {
		node.lhs, node.rhs = node.rhs, node.lhs
	}

	// ptr + num
	if node.lhs.ty.kind == tyPtr && node.rhs.ty.kind == tyInt {
		scale := newNode(ndMul, node.rhs, newNodeNum(8))
		if err := addType(scale); err != nil {
			return err
		}
		node.rhs = scale
		node.ty = node.lhs.ty
		return nil
	}

	return fmt.Errorf("invalid operands for +")
}

func typeSub(node *node) error {
	// num - num
	if node.lhs.ty.kind == tyInt && node.rhs.ty.kind == tyInt {
		node.ty = intType()
		return nil
	}

	// ptr - num
	if node.lhs.ty.kind == tyPtr && node.rhs.ty.kind == tyInt {
		scale := newNode(ndMul, node.rhs, newNodeNum(8))
		if err := addType(scale); err != nil {
			return err
		}
		node.rhs = scale
		node.ty = node.lhs.ty
		return nil
	}

	// ptr - ptr
	if node.lhs.ty.kind == tyPtr && node.rhs.ty.kind == tyPtr {
		sub := newNode(ndSub, node.lhs, node.rhs)
		sub.ty = intType()

		node.kind = ndDiv
		node.lhs = sub
		node.rhs = newNodeNum(8)
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
	case ndAssign:
		node.ty = node.lhs.ty
		return nil
	case ndMul, ndDiv, ndEq, ndNe, ndLt, ndLe, ndFuncall, ndNum:
		node.ty = intType()
		return nil
	case ndVar:
		node.ty = node.lvar.ty
		return nil
	case ndAddr:
		node.ty = pointerTo(node.lhs.ty)
		return nil
	case ndDeref:
		if node.lhs.ty.kind == tyPtr {
			node.ty = node.lhs.ty.base
			return nil
		} else {
			node.ty = intType()
			return nil
		}
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
