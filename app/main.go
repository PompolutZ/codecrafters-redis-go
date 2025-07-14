package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/radisa"
)

func main() {
	fmt.Println("Starting Radisa server...")

	dir := flag.String("dir", ".", "Directory to store RDB file")
	dbfilename := flag.String("dbfilename", "rdbfile", "Name of the RDB file")

	flag.Parse()

	server := radisa.NewRadisa(*dir, *dbfilename)
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
