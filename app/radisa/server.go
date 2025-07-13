package radisa

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const CRLF = "\r\n"

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
	
	for {
		if !scanner.Scan() {
			return
		}

		commandArrayLengthToken := scanner.Text()
		if !strings.HasPrefix(commandArrayLengthToken, "*") {
			fmt.Println("Invalid command format")
			conn.Write([]byte("-ERR invalid command format" + CRLF))
			return
		}

		commandArrayLength, err := strconv.Atoi(strings.TrimPrefix(commandArrayLengthToken, "*"))
		if err != nil {
			fmt.Println("Invalid array length")
			conn.Write([]byte("-ERR invalid array length" + CRLF))
			return
		}

		command, err := parseCommand(scanner)
		if err != nil {
			conn.Write([]byte("-ERR " + err.Error() + CRLF))
			return
		}

		switch command {
			case "PING": 
				conn.Write([]byte("+PONG" + CRLF))
				return
			case "ECHO": 
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					return
				}

				bulk := strings.Join(args, " ");
				conn.Write([]byte("$" + strconv.Itoa(len(bulk)) + CRLF + bulk + CRLF))
				return
			default:
				conn.Write([]byte("-ERR unknown command" + CRLF))
				return
		}
	}	
}

func parseCommand(scanner *bufio.Scanner) (string, error) {
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read length of the command")
	}

	if !strings.HasPrefix(scanner.Text(), "$") {
		return "", fmt.Errorf("invalid command format, expected '$' prefix followed by length")
	}

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read command")
	}

	return strings.TrimSpace(scanner.Text()), nil
}

func parseArguments(scanner *bufio.Scanner, argsLen int) ([]string, error) {
	args := make([]string, 0)
	for i := 1; i < argsLen; i++ {
		fmt.Printf("Parsing argument %d\n", i)
		if !scanner.Scan() {
			return args, fmt.Errorf("failed to read length of the argument")
		}

		if !strings.HasPrefix(scanner.Text(), "$") {
			return args, fmt.Errorf("invalid argument format, expected '$' prefix followed by length")
		}

		if !scanner.Scan() {
			return args, fmt.Errorf("failed to read argument")
		}

		args = append(args, strings.TrimSpace(scanner.Text()))
	}

	return args, nil
}	

