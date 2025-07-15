package radisa

import (
	"bufio"
	"fmt"
	"maps"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CRLF = "\r\n"
const NULL_BULK_STR = "$-1" + CRLF

// om du vet, du vet
type Radisa struct {
	Port int
	data map[string]string
	expires map[string]time.Time
	mu sync.RWMutex
	dir string
	dbfilename string
}

func NewRadisa(dir string, dbfilename string) *Radisa {
	file, err := os.ReadFile(dir + "/" + dbfilename)
	if err != nil {
		fmt.Printf("Error reading RDB file: %v\n", err)

		return &Radisa{
			Port: 6379, // Default Redis port
			data: make(map[string]string),
			expires: make(map[string]time.Time),
			mu:   sync.RWMutex{},
			dir: dir,
			dbfilename: dbfilename,
		}
	}

	parser := NewRDBParser(file)

	return &Radisa{
		Port: 6379, // Default Redis port
		data: parser.Parse(),
		expires: make(map[string]time.Time),
		mu:   sync.RWMutex{},
		dir: dir,
		dbfilename: dbfilename,
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
		
		go r.handleConnection(conn)
	}
}

func (r *Radisa)handleConnection(conn net.Conn) {
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
			continue
		}

		commandArrayLength, err := strconv.Atoi(strings.TrimPrefix(commandArrayLengthToken, "*"))
		if err != nil {
			fmt.Println("Invalid array length")
			conn.Write([]byte("-ERR invalid array length" + CRLF))
			continue
		}

		command, err := parseCommand(scanner)
		if err != nil {
			conn.Write([]byte("-ERR " + err.Error() + CRLF))
			continue
		}

		switch command {
			case "PING": 
				conn.Write([]byte("+PONG" + CRLF))
			case "ECHO": 
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					continue
				}

				bulk := strings.Join(args, " ");
				conn.Write([]byte("$" + strconv.Itoa(len(bulk)) + CRLF + bulk + CRLF))
			case "SET":
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					continue
				}

				if len(args) < 2 {
					conn.Write([]byte("-ERR wrong number of arguments for 'set' command" + CRLF))
					continue
				}

				r.mu.Lock()
				r.data[args[0]] = args[1]
				r.mu.Unlock()

				// We gonna handle only PX argument for now
				if len(args) > 2 && strings.ToUpper(args[2]) == "PX" {
					duration, err := strconv.Atoi(args[3])
					if err != nil {
						conn.Write([]byte("-ERR invalid duration for PX argument" + CRLF))
						continue	
					}		

					r.mu.Lock()				
					r.expires[args[0]] = time.Now().Add(time.Duration(duration) * time.Millisecond)
					r.mu.Unlock()	
				}
				conn.Write([]byte("+OK" + CRLF))
			case "GET":
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					continue
				}

				if len(args) < 1 {
					conn.Write([]byte("-ERR wrong number of arguments for 'get' command" + CRLF))
					continue
				}

				exp, exists := r.expires[args[0]]
				if exists && time.Now().After(exp) {
					conn.Write([]byte(NULL_BULK_STR))
					r.mu.Lock()
					delete(r.data, args[0])
					delete(r.expires, args[0])
					r.mu.Unlock()
					continue
				} else {
					r.mu.RLock()
					value, exists := r.data[args[0]]
					r.mu.RUnlock()	
	
					if !exists {
						conn.Write([]byte(NULL_BULK_STR))
						continue
					}
	
					conn.Write([]byte("$" + strconv.Itoa(len(value)) + CRLF + value + CRLF))
				}

			case "CONFIG": 
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					continue
				}

				if args[0] == "GET" && len(args) == 2 && args[1] == "dir" {
					conn.Write([]byte("*2" + CRLF + toBulkString("dir") + toBulkString(r.dir)))
					continue
				}

				if args[0] == "GET" && len(args) == 2 && args[1] == "dbfilename" {
					conn.Write([]byte("*2" + CRLF + toBulkString("dbfilename") + toBulkString(r.dbfilename)))
					continue
				}
			case "KEYS":
				args, err := parseArguments(scanner, commandArrayLength)
				if err != nil {
					conn.Write([]byte("-ERR " + err.Error() + CRLF))
					continue
				}

				if len(args) < 1 {
					conn.Write([]byte("-ERR wrong number of arguments for 'keys' command" + CRLF))
					continue
				}

				pattern := args[0]
				r.mu.RLock()
				keys := SearchKeys(pattern, slices.Collect(maps.Keys(r.data)))
				r.mu.RUnlock()

				bulkStrings := make([]string, len(keys))
				for i, key := range keys {
					bulkStrings[i] = toBulkString(key)
				}

				conn.Write([]byte("*" + strconv.Itoa(len(bulkStrings)) + CRLF + strings.Join(bulkStrings, "")))
			default:
				conn.Write([]byte("-ERR unknown command" + CRLF))
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

	return strings.ToUpper(strings.TrimSpace(scanner.Text())), nil
}

func parseArguments(scanner *bufio.Scanner, argsLen int) ([]string, error) {
	args := make([]string, 0)
	for i := 1; i < argsLen; i++ {
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

func toBulkString(value string) string {
	return "$" + strconv.Itoa(len(value)) + CRLF + value + CRLF
}

