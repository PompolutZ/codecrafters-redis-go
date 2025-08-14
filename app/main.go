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
	dbfilename := flag.String("dbfilename", "dump.rdb", "Name of the RDB file")
	port := flag.Int("port", 6379, "Port to run the server on")
	replicaof := flag.String("replicaof", "", "Start redis as replica of master")

	flag.Parse()

	var server *radisa.Radisa
	if *replicaof != "" {
		server = radisa.NewReplica(*dir, *dbfilename, *port, *replicaof)
	} else {
		server = radisa.NewRadisa(*dir, *dbfilename, *port)
	}

	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
