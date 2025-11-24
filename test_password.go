package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$VuGd0ekbW2lxZejZSZjKE.C548Fi9zIjx3XgfBKdKjZK53SW/C6OO"
	password := "password123"
	
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
	} else {
		fmt.Println("Password verified successfully!")
	}
}
