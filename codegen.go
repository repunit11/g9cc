package main

import (
	"fmt"
	"os"
)

var cntif int
var argregs64 = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var argregs32 = []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
var argregs8 = []string{"dil", "sil", "dl", "cl", "r8b", "r9b"}

func count() int {
	cntif++
	return cntif
}

func load(ty *ty) {
	if ty.kind == tyArray {
		return
	}
	if ty.size == 8 {
		fmt.Printf("	mov rax, [rax]\n")
	} else if ty.size == 4 {
		fmt.Printf("	mov eax, [rax]\n")
	} else if ty.size == 1 {
		fmt.Printf("	movsbq rax, [rax]\n")
	}
}

func store(ty *ty) {
	fmt.Printf("	pop rax\n")
	if ty.size == 8 {
		fmt.Printf("	mov [rax], rdi\n")
	} else if ty.size == 4 {
		fmt.Printf("	mov [rax], edi\n")
	} else if ty.size == 1 {
		fmt.Printf("	mov [rax], dil\n")
	}
}

func genExpr(node *node) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndVar:
		genAddr(node)
		fmt.Printf("	pop rax\n")
		load(node.ty)
		fmt.Printf("	push rax\n")
		return
	case ndAssign:
		genAddr(node.lhs)
		genExpr(node.rhs)
		fmt.Printf("	pop rdi\n")
		store(node.lhs.ty)
		fmt.Printf("	push rdi\n")
		return
	case ndFuncall: // TODO: 関数呼び出し前にRSPを16の倍数になるようにする
		for i := len(node.args) - 1; i >= 0; i-- {
			genExpr(node.args[i])
		}
		if len(node.args) > len(argregs64) {
			fmt.Fprintf(os.Stderr, "too many arguments: max %d\n", len(argregs64))
			os.Exit(1)
		}
		for i := 0; i < len(node.args); i++ {
			fmt.Printf("	pop %s\n", argregs64[i])
		}
		fmt.Printf("	call %s\n", node.funcname)
		fmt.Printf("	push rax\n")
		return
	case ndAddr:
		genAddr(node.lhs)
		return
	case ndDeref:
		genExpr(node.lhs)
		fmt.Printf("	pop rax\n")
		load(node.ty)
		fmt.Printf("	push rax\n")
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

func genFunc(funct *obj) {
	fmt.Printf("%s:\n", *funct.name)

	// プロローグ
	fmt.Printf("	push rbp\n")
	fmt.Printf("	mov rbp, rsp\n")
	fmt.Printf("	sub rsp, 208\n") // 208 = ('z' - 'a' + 1) * 8

	param := funct.params
	i := 0
	for param != nil {
		if param.ty.size == 4 {
			fmt.Printf("	mov [rbp - %d], %s\n", param.offset, argregs32[i])
		} else if param.ty.size == 8 {
			fmt.Printf("	mov [rbp - %d], %s\n", param.offset, argregs64[i])
		} else if param.ty.size == 1 {
			fmt.Printf("	mov [rbp - %d], %s\n", param.offset, argregs8[i])
		}
		param = param.next
		i++
	}

	// ASTの生成
	genStmt(funct.body)

	fmt.Printf("	mov rsp, rbp\n")
	fmt.Printf("	pop rbp\n")
	fmt.Printf("	ret\n")
}

// 左辺値のアドレス生成
func genAddr(node *node) {
	switch node.kind {
	case ndVar:
		if node.lvar.isLocal {
			offset := node.lvar.offset
			fmt.Printf("	mov rax, rbp\n")
			fmt.Printf("	sub rax, %d\n", offset)
			fmt.Printf("	push rax\n")
		} else {
			fmt.Printf("	lea rax, %s[rip]\n", *node.lvar.name)
			fmt.Printf("	push rax\n")
		}
		return
	case ndDeref:
		genExpr(node.lhs)
		return
	}

	fmt.Fprintf(os.Stderr, "not an lvalue")
}

func emitData(prog *obj) {
	fmt.Printf(".data\n")
	for v := prog; v != nil; v = v.next {
		if v.isFunction {
			continue
		}
		fmt.Printf(".global %s\n", *v.name)
		fmt.Printf("%s:\n", *v.name)
		fmt.Printf("	.zero %d\n", v.ty.size)
	}
}

func emitText(prog *obj) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".text\n")
	for v := prog; v != nil; v = v.next {
		if !v.isFunction {
			continue
		}
		fmt.Printf(".global %s\n", *v.name)
		genFunc(v)
	}
}

func codegen(prog *obj) {
	emitData(prog)
	emitText(prog)
}
