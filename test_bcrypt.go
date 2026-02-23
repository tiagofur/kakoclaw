package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) > 2 {
		// Verify mode: go run test_bcrypt.go hash password
		stored := os.Args[1]
		password := os.Args[2]
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(password))
		if err != nil {
			fmt.Printf("FAIL: %v\n", err)
		} else {
			fmt.Printf("SUCCESS!\n")
		}
		return
	}

	if len(os.Args) < 2 {
		password := "Passw0rd!2026"
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		fmt.Printf("Hash for '%s': %s\n", password, string(hash))
		return
	}
	password := os.Args[1]
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Printf("Hash for '%s': %s\n", password, string(hash))
}
