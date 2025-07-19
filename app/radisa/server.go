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

type Data struct {
	value string
	expire time.Time	
}

// om du vet, du vet
type Radisa struct {
	Port int
	data map[string]Data
	mu sync.RWMutex
	dir string
	dbfilename string
}

func NewRadisa(dir string, dbfilename string, port int) *Radisa {
	file, err := os.ReadFile(dir + "/" + dbfilename)
	if err != nil {
		fmt.Printf("Error reading RDB file: %v\n", err)

		return &Radisa{
			Port: port, // Default Redis port
			data: make(map[string]Data),
			mu:   sync.RWMutex{},
			dir: dir,
			dbfilename: dbfilename,
		}
	}

	parser := NewRDBParser(file)

	return &Radisa{
		Port: port,
		data: parser.Parse(),
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

func (r *Radisa) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)
	parser := NewRESPParser(scanner)
	
	for {
		// Parse RESP command
		cmd, err := parser.ParseCommand()
		if err != nil {
			// Check if it's EOF (client disconnected)
			if err.Error() == "failed to read command" {
				return
			}
			conn.Write(FormatError(err.Error()))
			continue
		}

		// Execute command and send response
		response := r.executeCommand(cmd)
		conn.Write(response)
	}	
}



// executeCommand processes a parsed command and returns the appropriate RESP response
func (r *Radisa) executeCommand(cmd *Command) []byte {
	switch cmd.Name {
	case "PING":
		return FormatSimpleString("PONG")

	case "ECHO":
		if len(cmd.Args) < 1 {
			return FormatError("wrong number of arguments for 'echo' command")
		}
		bulk := strings.Join(cmd.Args, " ")
		return FormatBulkString(bulk)

	case "SET":
		if len(cmd.Args) < 2 {
			return FormatError("wrong number of arguments for 'set' command")
		}

		key := cmd.Args[0]
		value := cmd.Args[1]
		expires := time.Time{}

		// Handle PX argument for expiry
		if len(cmd.Args) > 2 && strings.ToUpper(cmd.Args[2]) == "PX" {
			if len(cmd.Args) < 4 {
				return FormatError("invalid duration for PX argument")
			}
			duration, err := strconv.Atoi(cmd.Args[3])
			if err != nil {
				return FormatError("invalid duration for PX argument")
			}
			expires = time.Now().Add(time.Duration(duration) * time.Millisecond)
		}

		r.mu.Lock()
		r.data[key] = Data{
			value:  value,
			expire: expires,
		}
		r.mu.Unlock()

		return FormatSimpleString("OK")

	case "GET":
		if len(cmd.Args) < 1 {
			return FormatError("wrong number of arguments for 'get' command")
		}

		key := cmd.Args[0]
		r.mu.RLock()
		value, exists := r.data[key]
		r.mu.RUnlock()

		if !exists {
			return FormatNullBulkString()
		}

		if !value.expire.IsZero() && time.Now().After(value.expire) {
			r.mu.Lock()
			delete(r.data, key)
			r.mu.Unlock()
			return FormatNullBulkString()
		}

		return FormatBulkString(value.value)

	case "CONFIG":
		if len(cmd.Args) < 2 {
			return FormatError("wrong number of arguments for 'config' command")
		}

		if cmd.Args[0] == "GET" && cmd.Args[1] == "dir" {
			return FormatArray([]string{"dir", r.dir})
		}

		if cmd.Args[0] == "GET" && cmd.Args[1] == "dbfilename" {
			return FormatArray([]string{"dbfilename", r.dbfilename})
		}

		return FormatError("unknown config parameter")

	case "KEYS":
		if len(cmd.Args) < 1 {
			return FormatError("wrong number of arguments for 'keys' command")
		}

		pattern := cmd.Args[0]
		r.mu.RLock()
		keys := SearchKeys(pattern, slices.Collect(maps.Keys(r.data)))
		r.mu.RUnlock()

		return FormatArray(keys)

	case "INFO":
		return FormatBulkString("role:master")

	default:
		return FormatError("unknown command")
	}
}

