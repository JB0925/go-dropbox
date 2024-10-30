package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jbrink/go-dropbox/api"
)

func main() {
	var maxRequests int
	if len(os.Args) > 1 {
		m, err := strconv.Atoi(os.Args[1])
		if err != nil {
			panic("Provided improper arg for maxRequests for rate limiter.")
		}
		maxRequests = m
	}
	server := api.NewServer(maxRequests)
	err := api.StartServer(server)
	if err != nil {
		fmt.Printf("main.go::main - Error starting server: %v", err)
	}
}
