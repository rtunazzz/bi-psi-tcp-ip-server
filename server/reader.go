package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type RobotReader struct {
	Conn   net.Conn
	Buffer string
}

func (r *RobotReader) getMessage(maxLength int) (msg string, err error) {
	for {
		parts := strings.SplitN(r.Buffer, "\a\b", 2)

		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			msg = parts[0]

			// If we exceeded the max length of the message
			if len(msg) > (maxLength + 1) {
				log.Println("Maximum message length exceeded!")
				err = errors.New(SERVER_SYNTAX_ERROR)
				return
			}

			r.Buffer = parts[1]
			return
		}
		err = r.readSocketBuffer()
		if err != nil {
			log.Printf("Error occured during reading socket buffer: %s\n", err)
			return
		}
	}
}

func (r *RobotReader) readSocketBuffer() (err error) {
	// Set a deadline for reading. Read operation will fail if no data is received after deadline.
	// r.Conn.SetReadDeadline(time.Now().Add(TIMEOUT))

	recBuffer := make([]byte, BUFFER_SIZE)
	n, err := r.Conn.Read(recBuffer)
	if n == 0 || err != nil {
		log.Println("Failed to read connection:", err)
		// TODO Kouknout na zadani jestli tady vubec mam vracet nejaky error
		return err
	}
	if e, ok := err.(interface{ Timeout() bool }); ok && e.Timeout() {
		log.Println("Timeout error", e)
		// TODO Kouknout na zadani jestli tady vubec mam vracet nejaky error
		return err
	}

	// For debugging
	// r.Buffer = r.Buffer + strings.Replace(string(recBuffer[:n]), "\n", "", -1)

	// Convert the received buffer to string and add it to the main buffer
	r.Buffer = r.Buffer + string(recBuffer[:n])
	return nil
}

func (r *RobotReader) authenticate() (err error) {
	// Get robot's username
	username, err := r.getMessage(MAX_USERNAME_LEN)
	if err != nil {
		log.Printf("Error while getting robot's name: %s\n", err)
		return err
	}

	// TODO HANDLE AUTH

	log.Printf("[%s] Authenticating...\n", username)
	_, err = r.Conn.Write([]byte(SERVER_KEY_REQUEST))
	if err != nil {
		return err
	}

	recKeyIndexStr, err := r.getMessage(MAX_KEY_ID_LEN)
	if err != nil {
		log.Printf("[%s] Error while getting key id: %s\n", username, err)
		return err
	}

	// TODO Add check to make sure key index is valid (+ error handling)
	log.Printf("[%s] Looking for key index %s\n", username, recKeyIndexStr)
	serverKey, clientKey := authkeyLookup(recKeyIndexStr)
	log.Printf("[%s] Found serverKey: '%d' and clientKey: '%d'\n", username, serverKey, clientKey)

	hash := getHash(username)
	serverHash := (hash + serverKey) % 65536
	clientHash := (hash + clientKey) % 65536
	log.Printf("[%s] Sending server hash: '%d'\n", username, serverHash)
	_, err = r.Conn.Write([]byte(fmt.Sprint(serverHash, "\a\b")))
	if err != nil {
		return err
	}

	recClientHash, err := r.getMessage(MAX_CONFIRMATION_LEN)
	if err != nil {
		log.Printf("[%s] Error while receiving client hash: %s\n", username, err)
		return err
	}
	if len(recClientHash) > 5 {
		log.Printf("[%s] Client hash is too long. %s\n", username, recClientHash)
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	recClientHashInt, err := strconv.Atoi(recClientHash)
	if err != nil {
		log.Printf("[%s] Client hash is not a number. %s\n", username, recClientHash)
		return err
	}
	log.Printf("[%s] Recieved client hash '%s'.\n", username, recClientHash)
	log.Printf("[%s] Checking if client hashesh match ('%d' == '%s')\n", username, clientHash, recClientHash)
	if recClientHashInt == clientHash {
		log.Printf("[%s] Successfully authenticated.\n", username)
		_, err = r.Conn.Write([]byte(SERVER_OK))
		if err != nil {
			return err
		}
	} else {
		log.Printf("[%s] Failed to authenticate.\n", username)
		_, err = r.Conn.Write([]byte(SERVER_LOGIN_FAILED))
		if err != nil {
			return err
		}
	}
	return nil
}
