package main

import "fmt"

func main() {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("	mov rax, %d\n", 2)
	fmt.Printf("	add rax, %d\n", 2)
	fmt.Printf("	ret\n")
}
