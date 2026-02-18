package main

type typekind int

const (
	tyInt typekind = iota
	tyPtr
	tyArray
)

type ty struct {
	kind     typekind
	base     *ty
	name     *token
	size     int
	arrayLen int
}

func pointerTo(base *ty) *ty {
	ty := new(ty)
	ty.kind = tyPtr
	ty.base = base
	ty.size = 8
	return ty
}

func arrayOf(base *ty, len int) *ty {
	ty := &ty{
		kind:     tyArray,
		size:     base.size * len,
		base:     base,
		arrayLen: len,
	}
	return ty
}
