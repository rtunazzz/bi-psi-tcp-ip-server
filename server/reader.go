package server

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"
)

type Robot struct {
	Conn      net.Conn
	Buffer    string
	Username  string
	coors     *Coordinate
	prevCoors *Coordinate
	Direction Direction
}

// Gets a message from the Buffer property and returns it
func (r *Robot) getMessage(maxLength int) (msg string, err error) {
	for {
		parts := strings.SplitN(r.Buffer, "\a\b", 2)

		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			msg = parts[0]
			if msg == CLIENT_RECHARGING {
				err = r.recharge()
				if err != nil {
					return
				}
				continue
			}

			r.Buffer = parts[1]
			return
		} else if len(r.Buffer) > maxLength-1 {
			// If we exceeded the max length of the message
			log.Printf("Maximum message (%s) length exceeded! %d > %d\n", r.Buffer, len(r.Buffer), maxLength)
			err = errors.New(SERVER_SYNTAX_ERROR)
			return
		}

		err = r.readSocketBuffer(TIMEOUT)
		if err != nil {
			log.Printf("Error occured during reading socket buffer: %s\n", err)
			return
		}
	}
}

// Handles robot recharging
func (r *Robot) recharge() (err error) {
	for {
		err = r.readSocketBuffer(TIMEOUT_RECHARGING)
		if err != nil {
			log.Printf("[%s] [RECHARGING] Error occured during reading socket buffer: %s\n", r.Username, err)
			return
		}

		parts := strings.SplitN(r.Buffer, "\a\b", 2)
		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			// If we receive a message that isn't CLIENT_FULL_POWER
			if msg := parts[0]; msg != CLIENT_FULL_POWER {
				return errors.New(SERVER_LOGIC_ERROR)
			}
			return
		}
	}
}

// Reads the sockets buffer and saves its content into the Buffer property.
func (r *Robot) readSocketBuffer(timeout time.Duration) (err error) {
	// Set a deadline for reading. Read operation will fail if no data is received after deadline.
	r.Conn.SetReadDeadline(time.Now().Add(timeout))

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

	// log.Printf("Got new data (%d): %s", len(recBuffer[:n]), recBuffer[:n])
	// Convert the received buffer to string and add it to the main buffer
	r.Buffer = r.Buffer + string(recBuffer[:n])
	return nil
}

// Executed the command specified and waits for a response, then returns the response
func (r *Robot) executeCommandAndWaitForResponse(cmd string, maxMsgLength int) (res string, err error) {
	_, err = r.Conn.Write([]byte(cmd))
	if err != nil {
		return
	}
	res, err = r.getMessage(maxMsgLength)
	return
}
