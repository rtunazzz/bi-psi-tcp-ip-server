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
			fmt.Println(err)
			
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

func handleConnection(conn net.Conn) {
	defer func () {
		err := conn.Close();
		if err != nil {
			log.Println("Failed to close listener:", err)
		}
	}()
}
