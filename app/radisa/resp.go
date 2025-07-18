package radisa

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Command struct {
	Name string   
	Args []string 
}

type RESPParser struct {
	scanner *bufio.Scanner
}

func NewRESPParser(scanner *bufio.Scanner) *RESPParser {
	return &RESPParser{
		scanner: scanner,
	}
}

func (p *RESPParser) ParseCommand() (*Command, error) {
	// Read the array length token (e.g., "*3")
	if !p.scanner.Scan() {
		return nil, fmt.Errorf("failed to read command")
	}

	arrayLengthToken := p.scanner.Text()
	if !strings.HasPrefix(arrayLengthToken, "*") {
		return nil, fmt.Errorf("invalid command format, expected '*' prefix")
	}

	arrayLength, err := strconv.Atoi(strings.TrimPrefix(arrayLengthToken, "*"))
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %v", err)
	}

	if arrayLength < 1 {
		return nil, fmt.Errorf("command array must have at least one element")
	}

	commandName, err := p.parseBulkString()
	if err != nil {
		return nil, fmt.Errorf("failed to parse command name: %v", err)
	}

	args := make([]string, 0, arrayLength-1)
	for i := 1; i < arrayLength; i++ {
		arg, err := p.parseBulkString()
		if err != nil {
			return nil, fmt.Errorf("failed to parse argument %d: %v", i, err)
		}
		args = append(args, arg)
	}

	return &Command{
		Name: strings.ToUpper(strings.TrimSpace(commandName)),
		Args: args,
	}, nil
}

func (p *RESPParser) parseBulkString() (string, error) {
	// Read the length token (e.g., "$3")
	if !p.scanner.Scan() {
		return "", fmt.Errorf("failed to read bulk string length")
	}

	lengthToken := p.scanner.Text()
	if !strings.HasPrefix(lengthToken, "$") {
		return "", fmt.Errorf("invalid bulk string format, expected '$' prefix")
	}

	// Parse the length
	length, err := strconv.Atoi(strings.TrimPrefix(lengthToken, "$"))
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %v", err)
	}

	// Handle null bulk string
	if length == -1 {
		return "", nil
	}

	// Read the actual string content
	if !p.scanner.Scan() {
		return "", fmt.Errorf("failed to read bulk string content")
	}

	content := p.scanner.Text()
	return strings.TrimSpace(content), nil
}

// RESP Response Formatters

// FormatSimpleString formats a simple string response (e.g., "+OK\r\n")
func FormatSimpleString(s string) []byte {
	return []byte("+" + s + CRLF)
}

// FormatBulkString formats a bulk string response (e.g., "$3\r\nfoo\r\n")
func FormatBulkString(s string) []byte {
	if s == "" {
		return []byte(NULL_BULK_STR)
	}
	return []byte("$" + strconv.Itoa(len(s)) + CRLF + s + CRLF)
}

// FormatArray formats an array response (e.g., "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
func FormatArray(elements []string) []byte {
	if len(elements) == 0 {
		return []byte("*0" + CRLF)
	}

	result := "*" + strconv.Itoa(len(elements)) + CRLF
	for _, element := range elements {
		result += "$" + strconv.Itoa(len(element)) + CRLF + element + CRLF
	}
	return []byte(result)
}

// FormatError formats an error response (e.g., "-ERR message\r\n")
func FormatError(errMsg string) []byte {
	return []byte("-ERR " + errMsg + CRLF)
}

// FormatNullBulkString returns a null bulk string response
func FormatNullBulkString() []byte {
	return []byte(NULL_BULK_STR)
}
