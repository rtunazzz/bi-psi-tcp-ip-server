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
	TIMEOUT            = 1 * time.Second // Server i klient očekávají od protistrany odpověď po dobu tohoto intervalu.
	TIMEOUT_RECHARGING = 5 * time.Second // Časový interval, během kterého musí robot dokončit dobíjení.

	// Constatnt Server messages
	SERVER_MOVE                   = "102 MOVE\\a\\b"             //	Příkaz pro pohyb o jedno pole vpřed
	SERVER_TURN_LEFT              = "103 TURN LEFT\\a\\b"        //	Příkaz pro otočení doleva
	SERVER_TURN_RIGHT             = "104 TURN RIGHT\\a\\b"       //	Příkaz pro otočení doprava
	SERVER_PICK_UP                = "105 GET MESSAGE\\a\\b"      //	Příkaz pro vyzvednutí zprávy
	SERVER_LOGOUT                 = "106 LOGOUT\\a\\b"           //	Příkaz pro ukončení spojení po úspěšném vyzvednutí zprávy
	SERVER_KEY_REQUEST            = "107 KEY REQUEST\\a\\b"      //	Žádost serveru o Key ID pro komunikaci
	SERVER_OK                     = "200 OK\\a\\b"               //	Kladné potvrzení
	SERVER_LOGIN_FAILED           = "300 LOGIN FAILED\\a\\b"     //	Nezdařená autentizace
	SERVER_SYNTAX_ERROR           = "301 SYNTAX ERROR\\a\\b"     //	Chybná syntaxe zprávy
	SERVER_LOGIC_ERROR            = "302 LOGIC ERROR\\a\\b"      //	Zpráva odeslaná ve špatné situaci
	SERVER_KEY_OUT_OF_RANGE_ERROR = "303 KEY OUT OF RANGE\\a\\b" // Key ID není v očekávaném rozsahu

	// Constatnt Client Messages
	CLIENT_RECHARGING = "RECHARGING\\a\\b" // Robot se začal dobíjet a přestal reagovat na zprávy.
	CLIENT_FULL_POWER = "FULL POWER\\a\\b" // Robot doplnil energii a opět příjímá příkazy.
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
	log.Printf("[%s] Handling a new connection...\n", conn.RemoteAddr().String())

	defer func() {
		log.Println("Closing connection...")
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close listener:", err)
		}
	}()

	r := RobotReader{Conn: conn}
	msg, err := r.getMessage(10)
	if err != nil {
		log.Printf("Error while getting a message: %s\n", err)
		return
	}
	log.Printf("Received a message: %s\n", msg)
}
