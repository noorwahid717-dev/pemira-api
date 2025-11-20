package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke"
	password := "password123"
	
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
	} else {
		fmt.Println("Password verified successfully!")
	}
}
