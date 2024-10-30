package main

import (
	"fmt"

	"github.com/jbrink/go-dropbox/api"
)

func main() {
	server := api.NewServer()
	err := api.StartServer(server)
	if err != nil {
		fmt.Printf("main.go::main - Error starting server: %v", err)
	}
}
