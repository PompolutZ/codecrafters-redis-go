package radisa

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// om du vet, du vet
type Radisa struct {
	Port int
}

func NewRadisa() *Radisa {
	return &Radisa{
		Port: 6379, // Default Redis port
	}
}

func (r *Radisa) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.Port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\r\n", r.Port)
		os.Exit(1)
	}
	
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		}
	}
}


