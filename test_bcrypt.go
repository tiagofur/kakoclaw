package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	hash := "$2a$10$OXS3eXsc/m3knDux1nRcX.dh62ZFhck2kfIa4mDWfg/NRGbC.iCQW"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		fmt.Println("✓ Password matches hash!")
	} else {
		fmt.Printf("✗ Password does NOT match hash: %v\n", err)
	}
}