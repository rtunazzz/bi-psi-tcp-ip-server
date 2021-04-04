package server

import (
	"errors"
	"log"
	"net"
	"strings"
)

type RobotReader struct {
	Conn   net.Conn
	Buffer string
}

func (r *RobotReader) getMessage(maxLength int) (msg string, err error) {
	for {
		// If we exceeded the max length of the buffer
		if len(r.Buffer) > (maxLength + 1) {
			log.Println("Maximum message length exceeded!")
			err = errors.New(SERVER_SYNTAX_ERROR)
			return
		}

		parts := strings.SplitN(r.Buffer, "\\a\\b", 2)

		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			msg = parts[0]
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
	// conn.SetReadDeadline(time.Now().Add(TIMEOUT))

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
	// strBuffer += strings.Replace(string(recBuffer[:n]), "\n", "", -1)

	// Convert the received buffer to string and add it to the main buffer
	r.Buffer = r.Buffer + string(recBuffer[:n])
	return nil
}
