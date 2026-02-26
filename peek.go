package main

import (
	"fmt"
	"os"
)
func main() {
	b, _ := os.ReadFile("pkg/storage/sqlite.go")
	s := string(b)
	fmt.Printf("%q\n", s[3700:4000]) // Look around the 143 range
}
