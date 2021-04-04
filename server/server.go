package server

// TODO: Find out if we really need to read bytes or not

import (
	"log"
	"net"
	"strings"
	"time"
)

const (
	BUFFER_SIZE        = 1024
	TIMEOUT            = 1 * time.Second
	TIMEOUT_RECHARGING = 5 * time.Second
)

// Starts a TCP listener
func StartListener() {
	network_type := "tcp"
	network_addr := ":4000"

	// Create a listener on the
	ln, err := net.Listen(network_type, network_addr)
	if err != nil {
		log.Fatal("Failed to create a listener:", err)
		return
	}

	// Close when done
	defer ln.Close()

	log.Printf("[%s] [%s] Initialized!", strings.ToUpper(network_type), network_addr)

	// Handle incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept an incoming connection:", err)

			// Close connection
			err := conn.Close()
			if err != nil {
				log.Println("Failed to close listener:", err)
			}
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Printf("[%s] Handling a new connection...", conn.RemoteAddr().String())

	defer func() {
		log.Println("Closing connection...")
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close listener:", err)
		}
	}()

	strBuffer := ""
	for {
		// Set a deadline for reading. Read operation will fail if no data is received after deadline.
		// conn.SetReadDeadline(time.Now().Add(TIMEOUT))

		recBuffer := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(recBuffer)
		if n == 0 || err != nil {
			log.Println("Failed to read connection:", err)
			return
		}
		if e, ok := err.(interface{ Timeout() bool }); ok && e.Timeout() {
			log.Println("Timeout error", e)
			return
		}

		// For debugging
		// strBuffer += strings.Replace(string(recBuffer[:n]), "\n", "", -1)

		// Convert the received buffer to string and add it to the main buffer
		strBuffer += string(recBuffer[:n])

		log.Printf("strBuffer:'%s'\n", strBuffer)

		msg, rest := parseMessage(strBuffer)
		if msg != "" {
			log.Printf("Message is:'%s'\n", msg)
			log.Printf("Left to read is:'%s'\n", rest)
			handleMessage(msg)
			strBuffer = rest
		}
	}
}

func parseMessage(s string) (msg, rest string) {
	parts := strings.SplitN(s, "\\a\\b", 2)
	if len(parts) != 2 {
		return "", s
	}
	msg = parts[0]
	rest = parts[1]
	return
}

func handleMessage(msg string) {

}
