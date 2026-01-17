package main

import (
	"fmt"
	"os"
)

func gen_expr(node *node) {
	if node.kind == ndNum {
		fmt.Printf("	push %d\n", node.val)
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

func codegen(node *node) {
	// アセンブリの前半部分の出力
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// ASTの生成
	for node != nil {
		gen_stmt(node)
		if node.next != nil {
			fmt.Printf("	pop rax\n")
		}
		node = node.next
	}

	fmt.Printf("	pop rax\n")
	fmt.Printf("	ret\n")
}
