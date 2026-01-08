//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Build() error {
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "g9cc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Test() error {
	if err := Build(); err != nil {
		return err
	}
	fmt.Println("Running tests...")
	cmd := exec.Command("bash", "test.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Clean() {
	fmt.Println("Cleaning...")
	_ = os.Remove("g9cc")

	patterns := []string{"*.o", "*~", "tmp*", "out", "out.s"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			_ = os.RemoveAll(match)
		}
	}
}
