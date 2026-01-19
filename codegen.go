package main

import (
	"fmt"
	"os"
)

func gen_expr(node *node) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndVar:
		gen_addr(node)
		fmt.Printf("	pop rax\n")
		fmt.Printf("	mov rax, [rax]\n")
		fmt.Printf("	push rax\n")
		return
	case ndAssign:
		gen_addr(node.lhs)
		gen_expr(node.rhs)
		fmt.Printf("	pop rdi\n")
		fmt.Printf("	pop rax\n")
		fmt.Printf("	mov [rax], rdi\n")
		fmt.Printf("	push rdi\n")
		return
	}

	gen_expr(node.lhs)
	gen_expr(node.rhs)

	fmt.Printf("	pop rdi\n")
	fmt.Printf("	pop rax\n")

	switch node.kind {
	case ndAdd:
		fmt.Printf("	add rax, rdi\n")
		break
	case ndSub:
		fmt.Printf("	sub rax, rdi\n")
		break
	case ndMul:
		fmt.Printf("	imul rax, rdi\n")
		break
	case ndDiv:
		fmt.Printf("	cqo\n")
		fmt.Printf("	idiv rdi\n")
		break
	case ndEq:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	sete al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndNe:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setne al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndLt:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setl al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	case ndLe:
		fmt.Printf("	cmp rax, rdi\n")
		fmt.Printf("	setle al\n")
		fmt.Printf("	movzb rax, al\n")
		break
	default:
		fmt.Fprintf(os.Stderr, "unexpected node kind")
		os.Exit(1)
	}
	fmt.Printf("	push rax\n")
}

func gen_stmt(node *node) {
	if node.kind == ndExprStmt {
		gen_expr(node.lhs)
		return
	}

	fmt.Fprintf(os.Stderr, "invalid statement")
}

func gen_addr(node *node) {
	if node.kind == ndVar {
		offset := int(node.name-'a'+1) * 8
		fmt.Printf("	mov rax, rbp\n")
		fmt.Printf("	sub rax, %d\n", offset)
		fmt.Printf("	push rax\n")
		return
	}

	fmt.Fprintf(os.Stderr, "not an lvalue")
}

func codegen(node *node) {
	// アセンブリの前半部分の出力
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// プロローグ
	fmt.Printf("	push rbp\n")
	fmt.Printf("	mov rbp, rsp\n")
	fmt.Printf("	sub rsp, 208\n") // 208 = ('z' - 'a' + 1) * 8

	// ASTの生成
	for node != nil {
		gen_stmt(node)
		fmt.Printf("	pop rax\n")
		node = node.next
	}

	fmt.Printf("	mov rsp, rbp\n")
	fmt.Printf("	pop rbp\n")
	fmt.Printf("	ret\n")
}
