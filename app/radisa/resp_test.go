package radisa

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestRESPParser_ParseCommand_PING(t *testing.T) {
	input := "*1\r\n$4\r\nPING\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "PING" {
		t.Errorf("Expected command name 'PING', got: %s", cmd.Name)
	}

	if len(cmd.Args) != 0 {
		t.Errorf("Expected 0 arguments, got: %d", len(cmd.Args))
	}
}

func TestRESPParser_ParseCommand_ECHO(t *testing.T) {
	input := "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "ECHO" {
		t.Errorf("Expected command name 'ECHO', got: %s", cmd.Name)
	}

	expectedArgs := []string{"hello"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v, got: %v", expectedArgs, cmd.Args)
	}
}

func TestRESPParser_ParseCommand_SET(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "SET" {
		t.Errorf("Expected command name 'SET', got: %s", cmd.Name)
	}

	expectedArgs := []string{"key", "value"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v, got: %v", expectedArgs, cmd.Args)
	}
}

func TestRESPParser_ParseCommand_SET_WithExpiry(t *testing.T) {
	input := "*5\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$2\r\nPX\r\n$4\r\n1000\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "SET" {
		t.Errorf("Expected command name 'SET', got: %s", cmd.Name)
	}

	expectedArgs := []string{"key", "value", "PX", "1000"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v, got: %v", expectedArgs, cmd.Args)
	}
}

func TestRESPParser_ParseCommand_GET(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "GET" {
		t.Errorf("Expected command name 'GET', got: %s", cmd.Name)
	}

	expectedArgs := []string{"key"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v, got: %v", expectedArgs, cmd.Args)
	}
}

func TestRESPParser_ParseCommand_InvalidArrayPrefix(t *testing.T) {
	input := "INVALID\r\n$4\r\nPING\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for invalid array prefix")
	}

	if !strings.Contains(err.Error(), "invalid command format") {
		t.Errorf("Expected 'invalid command format' error, got: %v", err)
	}
}

func TestRESPParser_ParseCommand_InvalidArrayLength(t *testing.T) {
	input := "*abc\r\n$4\r\nPING\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for invalid array length")
	}

	if !strings.Contains(err.Error(), "invalid array length") {
		t.Errorf("Expected 'invalid array length' error, got: %v", err)
	}
}

func TestRESPParser_ParseCommand_ZeroArrayLength(t *testing.T) {
	input := "*0\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for zero array length")
	}

	if !strings.Contains(err.Error(), "command array must have at least one element") {
		t.Errorf("Expected 'command array must have at least one element' error, got: %v", err)
	}
}

func TestRESPParser_ParseCommand_InvalidBulkStringPrefix(t *testing.T) {
	input := "*1\r\nINVALID\r\nPING\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for invalid bulk string prefix")
	}

	if !strings.Contains(err.Error(), "invalid bulk string format") {
		t.Errorf("Expected 'invalid bulk string format' error, got: %v", err)
	}
}

func TestRESPParser_ParseCommand_InvalidBulkStringLength(t *testing.T) {
	input := "*1\r\n$abc\r\nPING\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for invalid bulk string length")
	}

	if !strings.Contains(err.Error(), "invalid bulk string length") {
		t.Errorf("Expected 'invalid bulk string length' error, got: %v", err)
	}
}

func TestRESPParser_ParseCommand_NullBulkString(t *testing.T) {
	input := "*2\r\n$4\r\nECHO\r\n$-1\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "ECHO" {
		t.Errorf("Expected command name 'ECHO', got: %s", cmd.Name)
	}

	expectedArgs := []string{""}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v, got: %v", expectedArgs, cmd.Args)
	}
}

func TestRESPParser_ParseCommand_EmptyInput(t *testing.T) {
	input := ""
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	_, err := parser.ParseCommand()
	if err == nil {
		t.Fatal("Expected error for empty input")
	}

	if !strings.Contains(err.Error(), "failed to read command") {
		t.Errorf("Expected 'failed to read command' error, got: %v", err)
	}
}

// Response Formatter Tests

func TestFormatSimpleString(t *testing.T) {
	result := FormatSimpleString("OK")
	expected := "+OK\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatBulkString(t *testing.T) {
	result := FormatBulkString("hello")
	expected := "$5\r\nhello\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatBulkString_Empty(t *testing.T) {
	result := FormatBulkString("")
	expected := "$-1\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatArray(t *testing.T) {
	elements := []string{"foo", "bar"}
	result := FormatArray(elements)
	expected := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatArray_Empty(t *testing.T) {
	elements := []string{}
	result := FormatArray(elements)
	expected := "*0\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatArray_Single(t *testing.T) {
	elements := []string{"dir"}
	result := FormatArray(elements)
	expected := "*1\r\n$3\r\ndir\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatError(t *testing.T) {
	result := FormatError("unknown command")
	expected := "-ERR unknown command\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestFormatNullBulkString(t *testing.T) {
	result := FormatNullBulkString()
	expected := "$-1\r\n"

	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

// Integration Tests

func TestRESPParser_MultipleCommands(t *testing.T) {
	input := "*1\r\n$4\r\nPING\r\n*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	// Parse first command
	cmd1, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error for first command, got: %v", err)
	}
	if cmd1.Name != "PING" {
		t.Errorf("Expected first command 'PING', got: %s", cmd1.Name)
	}

	// Parse second command
	cmd2, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error for second command, got: %v", err)
	}
	if cmd2.Name != "ECHO" {
		t.Errorf("Expected second command 'ECHO', got: %s", cmd2.Name)
	}
	if len(cmd2.Args) != 1 || cmd2.Args[0] != "hello" {
		t.Errorf("Expected second command args ['hello'], got: %v", cmd2.Args)
	}
}

func TestRESPParser_CaseInsensitiveCommand(t *testing.T) {
	input := "*1\r\n$4\r\nping\r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.Name != "PING" {
		t.Errorf("Expected command name 'PING' (uppercase), got: %s", cmd.Name)
	}
}

func TestRESPParser_WhitespaceHandling(t *testing.T) {
	input := "*2\r\n$4\r\nECHO\r\n$7\r\n  hello  \r\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	parser := NewRESPParser(scanner)

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedArgs := []string{"hello"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("Expected args %v (trimmed), got: %v", expectedArgs, cmd.Args)
	}
}
