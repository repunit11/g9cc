package main

type typekind int

const (
	tyInt typekind = iota
	tyPtr
)

type ty struct {
	kind typekind
	base *ty
	name *token
	size int
}

func pointerTo(base *ty) *ty {
	ty := new(ty)
	ty.kind = tyPtr
	ty.base = base
	ty.size = 8
	return ty
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
