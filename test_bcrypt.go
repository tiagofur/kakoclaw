package kakoclaw
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)











}	}		fmt.Printf("✗ Password does NOT match hash: %v\n", err)	} else {		fmt.Println("✓ Password matches hash!")	if err == nil {	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))	password := "admin123"	hash := "$2a$10$OXS3eXsc/m3knDux1nRcX.dh62ZFhck2kfIa4mDWfg/NRGbC.iCQW"func main() {