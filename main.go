package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	network_type := "tcp"
	network_addr := ":4000"

	// Create a listener on the 
	listener, err := net.Listen(network_type, network_addr)
	if err != nil {
		log.Fatal("Failed to create a listener:", err)
		return
	}
	
	// Close when done
	defer listener.Close()
	
	log.Printf("[%s] [%s] Initialized!", strings.ToUpper(network_type), network_addr)

	// Handle incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			if err := conn.Close(); err != nil {
				log.Println("failed to close listener:", err)
			}
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())

		// go handleConn ...
	}
}
