package main

type typekind int

const (
	tyInt typekind = iota
	tyChar
	tyPtr
	tyArray
	tyFunc
)

type ty struct {
	kind     typekind
	base     *ty
	returnTy *ty
	name     *token
	size     int
	arrayLen int
}

func pointerTo(base *ty) *ty {
	ty := &ty{
		kind: tyPtr,
		base: base,
		size: 8,
	}
	return ty
}

func arrayOf(base *ty, len int) *ty {
	ty := &ty{
		kind:     tyArray,
		base:     base,
		size:     base.size * len,
		arrayLen: len,
	}
	return ty
}

func funcType(returnTy *ty) *ty {
	return &ty{
		kind:     tyFunc,
		returnTy: returnTy,
		size:     1,
	}
}
