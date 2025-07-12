package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/radisa"
)

func main() {
	fmt.Println("Starting Radisa server...")

	server := radisa.NewRadisa()
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
