package main

func sema(prog *function) error {
	for fn := prog; fn != nil; fn = fn.next {
		addType(fn.body)
	}
	return nil
}

func addType(node *node) {
	if node == nil {
		return
	}

	addType(node.lhs)
	addType(node.rhs)
	addType(node.cond)
	addType(node.then)
	addType(node.els)
	addType(node.init)
	addType(node.inc)

	switch node.kind {
	case ndAdd, ndSub, ndAssign:
		node.ty = node.lhs.ty
		return
	case ndMul, ndDiv, ndEq, ndNe, ndLt, ndLe, ndFuncall, ndNum:
		node.ty = &ty{kind: tyInt, size: 4}
		return
	case ndVar:
		node.ty = node.lvar.ty
		return
	case ndAddr:
		node.ty = pointerTo(node.lhs.ty)
		return
	case ndDeref:
		if node.lhs.ty.kind == tyPtr {
			node.ty = node.lhs.ty.base
			return
		} else {
			node.ty = &ty{kind: tyInt, size: 4}
			return
		}
	}
}
