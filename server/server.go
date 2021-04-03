package server

// TODO: Find out if we really need to read bytes or not

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strings"
	"time"
)

const (
	TIMEOUT = 1 * time.Second
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
			err := conn.Close();
			if err != nil {
				log.Println("Failed to close listener:", err)
			}
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())

		go handleConnection(conn)
	}
}


// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\a', '\b'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func handleConnection(conn net.Conn) {
	log.Printf("[%s] Handling a new connection...", conn.RemoteAddr().String())

	defer func () {
		log.Println("Closing connection...")
		err := conn.Close();
		if err != nil {
			log.Println("Failed to close listener:", err)
		}
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Split(ScanCRLF)
	for scanner.Scan() {
		// Set a deadline for reading. Read operation will fail if no data is received after deadline.
		// conn.SetReadDeadline(time.Now().Add(TIMEOUT))

		temp := string(scanner.Text());
		log.Printf("temp is: %s\n", temp)
		if temp == "STOP" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println("Error reading input:", err)
	}
}
