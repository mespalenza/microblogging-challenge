package main

import "fmt"

func main() {
	fmt.Print("hola tweet")
}

type Tweet struct {
	UserID  string `json:"user_id"`
	Id      string `json:"id"`
	Content string `json:"content"`
	Date    string `json:"date"`
}
