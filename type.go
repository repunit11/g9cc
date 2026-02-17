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
