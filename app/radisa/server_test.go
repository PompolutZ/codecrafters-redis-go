package radisa

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

// Helper function to create a test server
func createTestServer() *Radisa {
	return &Radisa{
		Port: 0, // Will be assigned by the OS
		data: make(map[string]Data),
		dir:  "/tmp",
		dbfilename: "test.rdb",
	}
}

// Helper function to send command and get response
func sendCommand(conn net.Conn, command string) (string, error) {
	// Send command
	_, err := conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	var response strings.Builder
	
	// For simple responses, we only need one line
	if scanner.Scan() {
		response.WriteString(scanner.Text())
		response.WriteString("\r\n")
	}
	
	return response.String(), nil
}

// Helper function to send command and get multi-line response
func sendCommandMultiLine(conn net.Conn, command string) (string, error) {
	// Send command
	_, err := conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	var response strings.Builder
	
	// Read first line to determine response type
	if !scanner.Scan() {
		return "", nil
	}
	
	firstLine := scanner.Text()
	response.WriteString(firstLine)
	response.WriteString("\r\n")
	
	// If it's an array, read the specified number of elements
	if strings.HasPrefix(firstLine, "*") {
		arrayLen := firstLine[1:]
		if arrayLen == "0" {
			return response.String(), nil
		}
		
		// Read array elements (each element has length line + content line)
		elementsToRead := 2 // For testing, we'll read a reasonable number of lines
		if arrayLen == "2" {
			elementsToRead = 4 // 2 elements * 2 lines each
		}
		
		for i := 0; i < elementsToRead && scanner.Scan(); i++ {
			response.WriteString(scanner.Text())
			response.WriteString("\r\n")
		}
	} else if strings.HasPrefix(firstLine, "$") && firstLine != "$-1" {
		// Bulk string - read the content line
		if scanner.Scan() {
			response.WriteString(scanner.Text())
			response.WriteString("\r\n")
		}
	}
	
	return response.String(), nil
}

func TestServer_PING_Command(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send PING command
	response, err := sendCommand(conn, "*1\r\n$4\r\nPING\r\n")
	if err != nil {
		t.Fatalf("Failed to send PING command: %v", err)
	}
	
	expected := "+PONG\r\n"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestServer_ECHO_Command(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send ECHO command
	response, err := sendCommandMultiLine(conn, "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")
	if err != nil {
		t.Fatalf("Failed to send ECHO command: %v", err)
	}
	
	expected := "$5\r\nhello\r\n"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestServer_SET_GET_Commands(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send SET command
	setResponse, err := sendCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	if err != nil {
		t.Fatalf("Failed to send SET command: %v", err)
	}
	
	expectedSet := "+OK\r\n"
	if setResponse != expectedSet {
		t.Errorf("SET response: expected %q, got %q", expectedSet, setResponse)
	}
	
	// Send GET command
	getResponse, err := sendCommandMultiLine(conn, "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n")
	if err != nil {
		t.Fatalf("Failed to send GET command: %v", err)
	}
	
	expectedGet := "$5\r\nvalue\r\n"
	if getResponse != expectedGet {
		t.Errorf("GET response: expected %q, got %q", expectedGet, getResponse)
	}
}

func TestServer_SET_With_Expiry(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send SET command with 100ms expiry
	setResponse, err := sendCommand(conn, "*5\r\n$3\r\nSET\r\n$7\r\ntestkey\r\n$9\r\ntestvalue\r\n$2\r\nPX\r\n$3\r\n100\r\n")
	if err != nil {
		t.Fatalf("Failed to send SET command: %v", err)
	}
	
	expectedSet := "+OK\r\n"
	if setResponse != expectedSet {
		t.Errorf("SET response: expected %q, got %q", expectedSet, setResponse)
	}
	
	// Immediately GET the value (should exist)
	getResponse1, err := sendCommandMultiLine(conn, "*2\r\n$3\r\nGET\r\n$7\r\ntestkey\r\n")
	if err != nil {
		t.Fatalf("Failed to send first GET command: %v", err)
	}
	
	expectedGet1 := "$9\r\ntestvalue\r\n"
	if getResponse1 != expectedGet1 {
		t.Errorf("First GET response: expected %q, got %q", expectedGet1, getResponse1)
	}
	
	// Wait for expiry
	time.Sleep(150 * time.Millisecond)
	
	// GET the value after expiry (should be null)
	getResponse2, err := sendCommand(conn, "*2\r\n$3\r\nGET\r\n$7\r\ntestkey\r\n")
	if err != nil {
		t.Fatalf("Failed to send second GET command: %v", err)
	}
	
	expectedGet2 := "$-1\r\n"
	if getResponse2 != expectedGet2 {
		t.Errorf("Second GET response: expected %q, got %q", expectedGet2, getResponse2)
	}
}

func TestServer_GET_NonExistent_Key(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send GET command for non-existent key
	response, err := sendCommand(conn, "*2\r\n$3\r\nGET\r\n$11\r\nnonexistent\r\n")
	if err != nil {
		t.Fatalf("Failed to send GET command: %v", err)
	}
	
	expected := "$-1\r\n"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestServer_CONFIG_Commands(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send CONFIG GET dir command
	response, err := sendCommandMultiLine(conn, "*3\r\n$6\r\nCONFIG\r\n$3\r\nGET\r\n$3\r\ndir\r\n")
	if err != nil {
		t.Fatalf("Failed to send CONFIG GET dir command: %v", err)
	}
	
	expectedStart := "*2\r\n$3\r\ndir\r\n$4\r\n/tmp\r\n"
	if response != expectedStart {
		t.Errorf("CONFIG GET dir response: expected %q, got %q", expectedStart, response)
	}
}

func TestServer_KEYS_Command(t *testing.T) {
	server := createTestServer()
	
	// Pre-populate some test data
	server.data["foo"] = Data{value: "bar", expire: time.Time{}}
	server.data["test"] = Data{value: "value", expire: time.Time{}}
	server.data["another"] = Data{value: "data", expire: time.Time{}}
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send KEYS * command
	response, err := sendCommandMultiLine(conn, "*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n")
	if err != nil {
		t.Fatalf("Failed to send KEYS command: %v", err)
	}
	
	// Response should be an array with 3 elements
	if !strings.HasPrefix(response, "*3\r\n") {
		t.Errorf("KEYS response should start with '*3\\r\\n', got: %q", response)
	}
}

func TestServer_Invalid_Command(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send invalid command
	response, err := sendCommand(conn, "*1\r\n$7\r\nINVALID\r\n")
	if err != nil {
		t.Fatalf("Failed to send invalid command: %v", err)
	}
	
	expected := "-ERR unknown command\r\n"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestServer_Wrong_Number_Of_Arguments(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send SET command with missing arguments
	response, err := sendCommand(conn, "*2\r\n$3\r\nSET\r\n$3\r\nkey\r\n")
	if err != nil {
		t.Fatalf("Failed to send SET command: %v", err)
	}
	
	expected := "-ERR wrong number of arguments for 'set' command\r\n"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestServer_Multiple_Commands_Same_Connection(t *testing.T) {
	server := createTestServer()
	
	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()
	
	// Connect to server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	
	// Send multiple commands on same connection
	commands := []struct {
		command  string
		expected string
	}{
		{"*1\r\n$4\r\nPING\r\n", "+PONG\r\n"},
		{"*3\r\n$3\r\nSET\r\n$4\r\ntest\r\n$4\r\ndata\r\n", "+OK\r\n"},
		{"*2\r\n$3\r\nGET\r\n$4\r\ntest\r\n", "$4\r\ndata\r\n"},
	}
	
	for i, cmd := range commands {
		var response string
		var err error
		
		if i == 2 { // GET command needs multi-line response
			response, err = sendCommandMultiLine(conn, cmd.command)
		} else {
			response, err = sendCommand(conn, cmd.command)
		}
		
		if err != nil {
			t.Fatalf("Failed to send command %d: %v", i, err)
		}
		
		if response != cmd.expected {
			t.Errorf("Command %d: expected %q, got %q", i, cmd.expected, response)
		}
	}
}
