package radisa

import (
	"encoding/binary"
	"fmt"
	"time"
)

type RDBParser struct {
	keyVals map[string]Data
	data []byte
	pos  int
}

func NewRDBParser(data []byte) *RDBParser {
	return &RDBParser{
		keyVals: make(map[string]Data),
		data: data,
		pos:  0,
	}
}

func (p *RDBParser) Parse() map[string]Data {
	// Parse header
	if !p.parseHeader() {
		return p.keyVals
	}

	// Parse metadata section
	p.parseMetadata()

	// Parse database section
	p.parseDatabase()

	// Parse end of file
	p.parseEOF()

	return p.keyVals
}

func (p *RDBParser) parseHeader() bool {
	if len(p.data) < 9 {
		fmt.Println("File too small")
		return false
	}

	magic := string(p.data[0:5])
	version := string(p.data[5:9])
	
	if magic != "REDIS" {
		fmt.Printf("Invalid magic string: %s\n", magic)
		return false
	}

	fmt.Printf("Magic: %s\n", magic)
	fmt.Printf("Version: %s\n", version)
	
	p.pos = 9
	return true
}

func (p *RDBParser) parseMetadata() {
	fmt.Println("\n=== Metadata Section ===")
	
	for p.pos < len(p.data) {
		if p.data[p.pos] != 0xFA {
			break // Not a metadata subsection
		}
		
		p.pos++ // Skip 0xFA
		
		// Read metadata name
		name := p.readString()
		// Read metadata value
		value := p.readString()
		
		fmt.Printf("Metadata: %s = %s\n", name, value)
	}
}

func (p *RDBParser) parseDatabase() {
	fmt.Println("\n=== Database Section ===")
	
	for p.pos < len(p.data) {
		if p.data[p.pos] == 0xFF {
			break // End of file marker
		}
		
		if p.data[p.pos] == 0xFE {
			// Database subsection
			p.pos++ // Skip 0xFE
			
			dbIndex := p.readSize()
			fmt.Printf("Database index: %d\n", dbIndex)
			
			// Check for hash table size info
			if p.pos < len(p.data) && p.data[p.pos] == 0xFB {
				p.pos++ // Skip 0xFB
				
				hashTableSize := p.readSize()
				expireTableSize := p.readSize()
				
				fmt.Printf("Hash table size: %d\n", hashTableSize)
				fmt.Printf("Expire table size: %d\n", expireTableSize)
			}
			
			// Parse key-value pairs
			p.parseKeyValuePairs()
		} else {
			break
		}
	}
}

func (p *RDBParser) parseKeyValuePairs() {
	fmt.Println("\n--- Key-Value Pairs ---")
	
	for p.pos < len(p.data) {
		if p.data[p.pos] == 0xFF {
			break // End of file
		}
		
		var expireTime int64 = -1
		var expireType string = ""
		
		// Check for expire information
		if p.data[p.pos] == 0xFD {
			// Expire in seconds
			p.pos++
			expireTime = int64(binary.LittleEndian.Uint32(p.data[p.pos:p.pos+4]))
			expireType = "seconds"
			p.pos += 4
		} else if p.data[p.pos] == 0xFC {
			// Expire in milliseconds
			p.pos++
			expireTime = int64(binary.LittleEndian.Uint64(p.data[p.pos:p.pos+8]))
			expireType = "milliseconds"
			p.pos += 8
		}
		
		// Read value type
		valueType := p.data[p.pos]
		p.pos++
		
		// Read key
		key := p.readString()
		
		// Read value based on type
		var value interface{}
		switch valueType {
		case 0x00: // String
			value = p.readString()
		case 0x01: // List
			value = p.readList()
		case 0x02: // Set
			value = p.readSet()
		case 0x04: // Hash
			value = p.readHash()
		default:
			fmt.Printf("Unknown value type: 0x%02X\n", valueType)
			return
		}

		var expire time.Time
		if expireTime != -1 {
			if expireType == "seconds" {
				expire = time.Unix(expireTime, 0)
			} else {
				expire = time.UnixMilli(expireTime)
			}
		}

		p.keyVals[key] = Data{
			value: fmt.Sprintf("%v", value),
			expire: expire,
		}
		
		fmt.Printf("Key: %s\n", key)
		fmt.Printf("Value: %v\n", value)
		fmt.Printf("Type: %s\n", p.getValueTypeName(valueType))
		
		if expireTime != -1 {
			fmt.Printf("Expires: %d (%s)\n", expireTime, expireType)
		}
		fmt.Println()
	}
}

func (p *RDBParser) readSize() uint64 {
	if p.pos >= len(p.data) {
		return 0
	}
	
	first := p.data[p.pos]
	p.pos++
	
	switch (first & 0xC0) >> 6 {
	case 0: // 00: 6-bit size
		return uint64(first & 0x3F)
	case 1: // 01: 14-bit size
		if p.pos >= len(p.data) {
			return 0
		}
		second := p.data[p.pos]
		p.pos++
		return uint64(((first & 0x3F) << 8) | second)
	case 2: // 10: 32-bit size
		if p.pos+4 > len(p.data) {
			return 0
		}
		size := binary.BigEndian.Uint32(p.data[p.pos : p.pos+4])
		p.pos += 4
		return uint64(size)
	case 3: // 11: Special string encoding
		return uint64(first & 0x3F)
	}
	return 0
}

func (p *RDBParser) readString() string {
	if p.pos >= len(p.data) {
		return ""
	}
	
	first := p.data[p.pos]
	
	// Check if it's special string encoding (first two bits are 11)
	if (first & 0xC0) == 0xC0 {
		p.pos++
		return p.readSpecialString(first & 0x3F)
	}
	
	// Regular string encoding
	length := p.readSize()
	if p.pos+int(length) > len(p.data) {
		return ""
	}
	
	str := string(p.data[p.pos : p.pos+int(length)])
	p.pos += int(length)
	return str
}

func (p *RDBParser) readSpecialString(encoding byte) string {
	switch encoding {
	case 0: // 8-bit integer
		if p.pos >= len(p.data) {
			return ""
		}
		val := p.data[p.pos]
		p.pos++
		return fmt.Sprintf("%d", val)
	case 1: // 16-bit integer
		if p.pos+2 > len(p.data) {
			return ""
		}
		val := binary.LittleEndian.Uint16(p.data[p.pos : p.pos+2])
		p.pos += 2
		return fmt.Sprintf("%d", val)
	case 2: // 32-bit integer
		if p.pos+4 > len(p.data) {
			return ""
		}
		val := binary.LittleEndian.Uint32(p.data[p.pos : p.pos+4])
		p.pos += 4
		return fmt.Sprintf("%d", val)
	case 3: // LZF compressed (not implemented)
		return "[LZF compressed - not implemented]"
	default:
		return fmt.Sprintf("[Unknown encoding: %d]", encoding)
	}
}

func (p *RDBParser) readList() []string {
	length := p.readSize()
	list := make([]string, length)
	
	for i := uint64(0); i < length; i++ {
		list[i] = p.readString()
	}
	
	return list
}

func (p *RDBParser) readSet() []string {
	length := p.readSize()
	set := make([]string, length)
	
	for i := uint64(0); i < length; i++ {
		set[i] = p.readString()
	}
	
	return set
}

func (p *RDBParser) readHash() map[string]string {
	length := p.readSize()
	hash := make(map[string]string)
	
	for i := uint64(0); i < length; i++ {
		key := p.readString()
		value := p.readString()
		hash[key] = value
	}
	
	return hash
}

func (p *RDBParser) getValueTypeName(valueType byte) string {
	switch valueType {
	case 0x00:
		return "String"
	case 0x01:
		return "List"
	case 0x02:
		return "Set"
	case 0x04:
		return "Hash"
	default:
		return fmt.Sprintf("Unknown (0x%02X)", valueType)
	}
}

func (p *RDBParser) parseEOF() {
	fmt.Println("\n=== End of File ===")
	
	if p.pos < len(p.data) && p.data[p.pos] == 0xFF {
		p.pos++ // Skip 0xFF
		
		if p.pos+8 <= len(p.data) {
			checksum := p.data[p.pos : p.pos+8]
			fmt.Printf("Checksum: %02X\n", checksum)
		}
	}
}