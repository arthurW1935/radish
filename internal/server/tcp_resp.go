package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"radish/internal/cache"
	"strconv"
	"strings"
)

type TCPServer struct {
	cache *cache.CacheManager
}

func NewTCPServer(c *cache.CacheManager) *TCPServer {
	return &TCPServer{cache: c}
}

func (s *TCPServer) Start() {
	listener, err := net.Listen("tcp", "0.0.0.0:7171")
	if err != nil {
		log.Fatal("Failed to start TCP server:", err)
	}
	defer listener.Close()

	log.Println("TCP RESP server running on port 7171...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read the first line (*<numArgs>)
		header, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Client disconnected:", err)
			return
		}
		header = strings.TrimSpace(header)

		// Validate RESP array format
		if len(header) < 2 || header[0] != '*' {
			conn.Write([]byte("-ERR Invalid RESP format\r\n"))
			continue
		}

		// Get argument count
		argCount, err := strconv.Atoi(header[1:])
		if err != nil || argCount < 1 {
			conn.Write([]byte("-ERR Invalid argument count\r\n"))
			continue
		}

		// Read and process arguments **directly** without storing in a slice
		var key, value string
		var command string

		for i := 0; i < argCount; i++ {
			// Read bulk string length (e.g., "$3\r\n")
			lengthLine, err := reader.ReadString('\n')
			if err != nil {
				conn.Write([]byte("-ERR Malformed RESP request\r\n"))
				return
			}
			lengthLine = strings.TrimSpace(lengthLine)

			// Validate RESP bulk string format
			if len(lengthLine) < 2 || lengthLine[0] != '$' {
				conn.Write([]byte("-ERR Invalid bulk string format\r\n"))
				return
			}

			// Get argument length
			argLen, err := strconv.Atoi(lengthLine[1:])
			if err != nil || argLen < 0 {
				conn.Write([]byte("-ERR Invalid bulk string length\r\n"))
				return
			}

			// Read actual argument (argLen bytes + \r\n)
			arg := make([]byte, argLen+2) // Include \r\n
			_, err = reader.Read(arg)
			if err != nil {
				conn.Write([]byte("-ERR Error reading argument\r\n"))
				return
			}

			// Remove \r\n
			parsedArg := string(arg[:argLen])

			// Directly process arguments based on position
			if i == 0 {
				command = strings.ToUpper(parsedArg) // First argument is the command
			} else if i == 1 {
				key = parsedArg // Second argument is the key
			} else if i == 2 {
				value = parsedArg // Third argument is the value (if present)
			}
		}

		// Execute command immediately without extra allocations
		var response string
		switch command {
		case "PUT":
			if key == "" || value == "" {
				response = "-ERR PUT requires key and value\r\n"
			} else {
				s.cache.Put(key, value)
				response = "+OK\r\n"
			}

		case "GET":
			if key == "" {
				response = "-ERR GET requires a key\r\n"
			} else {
				val, exists := s.cache.Get(key)
				if exists {
					response = fmt.Sprintf("$%d\r\n%s\r\n", len(val), val) // RESP bulk string response
				} else {
					response = "$-1\r\n" // RESP nil response if key is missing
				}
			}

		default:
			response = "-ERR Unknown command\r\n"
		}

		// Send response back
		conn.Write([]byte(response))
	}
}


// func (s *TCPServer) handleConnection(conn net.Conn) {
// 	defer conn.Close()
// 	reader := bufio.NewReader(conn)

// 	for {
// 		// Read the first line (should be *<numArgs>)
// 		header, err := reader.ReadString('\n')
// 		if err != nil {
// 			log.Println("Client disconnected:", err)
// 			return
// 		}
// 		header = strings.TrimSpace(header)

// 		// Ensure it's a RESP array
// 		if len(header) < 2 || header[0] != '*' {
// 			log.Println("Invalid RESP format:", header)
// 			conn.Write([]byte("-ERR Invalid RESP format\r\n"))
// 			continue
// 		}

// 		// Parse number of arguments
// 		argCount, err := strconv.Atoi(header[1:])
// 		if err != nil || argCount < 1 {
// 			conn.Write([]byte("-ERR Invalid argument count\r\n"))
// 			continue
// 		}

// 		args := make([]string, 0, argCount)
// 		for i := 0; i < argCount; i++ {
// 			// Read bulk string length (e.g., "$3")
// 			lengthLine, err := reader.ReadString('\n')
// 			if err != nil {
// 				log.Println("Error reading argument length:", err)
// 				conn.Write([]byte("-ERR Malformed RESP request\r\n"))
// 				return
// 			}
// 			lengthLine = strings.TrimSpace(lengthLine)

// 			// Ensure it's a valid RESP bulk string
// 			if len(lengthLine) < 2 || lengthLine[0] != '$' {
// 				conn.Write([]byte("-ERR Invalid bulk string format\r\n"))
// 				return
// 			}

// 			// Get argument length
// 			argLen, err := strconv.Atoi(lengthLine[1:])
// 			if err != nil || argLen < 0 {
// 				conn.Write([]byte("-ERR Invalid bulk string length\r\n"))
// 				return
// 			}

// 			// Read the actual argument (argLen bytes + \r\n)
// 			arg := make([]byte, argLen+2) // Include \r\n
// 			_, err = reader.Read(arg)
// 			if err != nil {
// 				log.Println("Error reading argument data:", err)
// 				conn.Write([]byte("-ERR Error reading argument\r\n"))
// 				return
// 			}

// 			// Store the argument, trimming \r\n
// 			args = append(args, string(arg[:argLen]))
// 		}

// 		// Process the full command now
// 		response := s.processRESP(args)
// 		conn.Write([]byte(response))
// 	}
// }


// // processRESP parses RESP protocol and executes commands
// // func (s *TCPServer) processRESP(input string) string {
// // 	lines := strings.Split(strings.TrimSpace(input), "\r\n")
// // 	if len(lines) < 1 || lines[0][0] != '*' {
// // 		log.Println("Invalid RESP format:", input)
// // 		return "-ERR Invalid RESP format\r\n"
// // 	}

// // 	// Read number of arguments
// // 	argCount, err := strconv.Atoi(lines[0][1:])
// // 	if err != nil || argCount < 1 {
// // 		return "-ERR Invalid argument count\r\n"
// // 	}

// // 	if len(lines) < (argCount*2 + 1) {
// // 		return "-ERR Malformed RESP request\r\n"
// // 	}

// // 	// Extract command and arguments
// // 	args := make([]string, 0, argCount)
// // 	for i := 1; i < len(lines); i += 2 {
// // 		args = append(args, lines[i+1])
// // 	}

// // 	command := strings.ToUpper(args[0])

// // 	switch command {
// // 	case "PUT":
// // 		if len(args) < 3 {
// // 			return "-ERR PUT requires key and value\r\n"
// // 		}
// // 		key, value := args[1], args[2]
// // 		s.cache.Put(key, value)
// // 		return "+OK\r\n"

// // 	case "GET":
// // 		if len(args) < 2 {
// // 			return "-ERR GET requires a key\r\n"
// // 		}
// // 		key := args[1]
// // 		value, exists := s.cache.Get(key)
// // 		if exists {
// // 			return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
// // 		}
// // 		return "$-1\r\n" // Redis-style nil response

// // 	default:
// // 		return "-ERR Unknown command\r\n"
// // 	}
// // }

// func (s *TCPServer) processRESP(args []string) string {
// 	if len(args) < 1 {
// 		return "-ERR Missing command\r\n"
// 	}

// 	command := strings.ToUpper(args[0])

// 	switch command {
// 	case "PUT":
// 		if len(args) < 3 {
// 			return "-ERR PUT requires key and value\r\n"
// 		}
// 		key, value := args[1], args[2]
// 		s.cache.Put(key, value)
// 		return "+OK\r\n"

// 	case "GET":
// 		if len(args) < 2 {
// 			return "-ERR GET requires a key\r\n"
// 		}
// 		key := args[1]
// 		value, exists := s.cache.Get(key)
// 		if exists {
// 			return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value) // Proper RESP bulk string response
// 		}
// 		return "$-1\r\n" // RESP nil response if key is not found

// 	default:
// 		return "-ERR Unknown command\r\n"
// 	}
// }
