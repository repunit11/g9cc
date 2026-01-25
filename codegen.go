package main

import (
	"fmt"
	"os"
)

var cntif int

func count() int {
	cntif++
	return cntif
}
func genExpr(node *node) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndVar:
		genAddr(node)
		fmt.Printf("	pop rax\n")
		fmt.Printf("	mov rax, [rax]\n")
		fmt.Printf("	push rax\n")
		return
	case ndAssign:
		genAddr(node.lhs)
		genExpr(node.rhs)
		fmt.Printf("	pop rdi\n")
		fmt.Printf("	pop rax\n")
		fmt.Printf("	mov [rax], rdi\n")
		fmt.Printf("	push rdi\n")
		return
	}

	genExpr(node.lhs)
	genExpr(node.rhs)

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

// 文のコード生成
func genStmt(node *node) {
	switch node.kind {
	case ndExprStmt:
		genExpr(node.lhs)
		fmt.Printf("	pop rax\n")
		return
	case ndReturn:
		genExpr(node.lhs)
		fmt.Printf("	pop rax\n")
		fmt.Printf("	mov rsp, rbp\n")
		fmt.Printf("	pop rbp\n")
		fmt.Printf("	ret\n")
		return
	case ndIf:
		cnt := count()
		genExpr(node.cond)
		fmt.Printf("	pop rax\n")
		fmt.Printf("	cmp rax, 0\n")
		fmt.Printf("	je .Lelse%d\n", cnt)
		genStmt(node.then)
		fmt.Printf("	jmp .Lend%d\n", cnt)
		fmt.Printf(".Lelse%d:\n", cnt)
		if node.els != nil {
			genStmt(node.els)
		}
		fmt.Printf(".Lend%d:\n", cnt)
		return
	case ndWhile:
		cnt := count()
		fmt.Printf(".Lbegin%d:\n", cnt)
		genExpr(node.lhs)
		fmt.Printf("	pop rax\n")
		fmt.Printf("	cmp rax, 0\n")
		fmt.Printf("	je	.Lend%d\n", cnt)
		genStmt(node.rhs)
		fmt.Printf("	jmp	.Lbegin%d\n", cnt)
		fmt.Printf(".Lend%d:\n", cnt)
		return
	case ndFor:
		cnt := count()
		if node.init != nil {
			genExpr(node.init)
			fmt.Printf("	pop rax\n")
		}
		fmt.Printf(".Lbegin%d:\n", cnt)
		if node.cond != nil {
			genExpr(node.cond)
			fmt.Printf("	pop rax\n")
			fmt.Printf("	cmp rax, 0\n")
			fmt.Printf("	je .Lend%d\n", cnt)
		}
		if node.then != nil {
			genStmt(node.then)
		}
		if node.inc != nil {
			genExpr(node.inc)
			fmt.Printf("	pop rax\n")
		}
		fmt.Printf("	jmp .Lbegin%d\n", cnt)
		fmt.Printf(".Lend%d:\n", cnt)
		return
	case ndBlock:
		n := node.lhs
		for n != nil {
			genStmt(n)
			n = n.next
		}
		return
	default:
		fmt.Fprintf(os.Stderr, "invalid statement")
		os.Exit(1)
	}
}

// 左辺値のアドレス生成
func genAddr(node *node) {
	if node.kind == ndVar {
		offset := node.offset
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
		genStmt(node)
		node = node.next
	}

	fmt.Printf("	mov rsp, rbp\n")
	fmt.Printf("	pop rbp\n")
	fmt.Printf("	ret\n")
}
