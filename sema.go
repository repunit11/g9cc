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

func walk(nodes ...*node) error {
	for _, n := range nodes {
		if err := addType(n); err != nil {
			return err
		}
	}
	return nil
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
		// num + num
		if node.lhs.ty.kind == tyInt && node.rhs.ty.kind == tyInt {
			node.ty = &ty{
				kind: tyInt,
				size: 4,
			}
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
	case ndSub:
		// num - num
		if node.lhs.ty.kind == tyInt && node.rhs.ty.kind == tyInt {
			node.ty = &ty{kind: tyInt, size: 4}
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
			basesize := 8
			sub := newNode(ndSub, node.lhs, node.rhs)
			sub.ty = &ty{kind: tyInt, size: 4}

			node.kind = ndDiv
			node.lhs = sub
			node.rhs = newNodeNum(basesize)
			node.ty = &ty{kind: tyInt, size: 4}
			return nil
		}
		return fmt.Errorf("invalid operands for -")
	case ndAssign:
		node.ty = node.lhs.ty
		return nil
	case ndMul, ndDiv, ndEq, ndNe, ndLt, ndLe, ndFuncall, ndNum:
		node.ty = &ty{kind: tyInt, size: 4}
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
			node.ty = &ty{kind: tyInt, size: 4}
			return nil
		}
	case ndSizeof:
		node.ty = &ty{
			kind: tyInt,
			base: nil,
			name: nil,
			size: 4,
		}
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
