package main

// ======================================  DISCLAIMER  ======================================
// Highly recommend reading through the code via the repository linked below.
// https://gitlab.fit.cvut.cz/hnatartu/osy-tcpip-server
// ==========================================================================================

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"gitlab.fit.cvut.cz/hnatartu/osy-tcpip-server/server"
)

const (
	// All are including the '\a' and '\b' characters
	MAX_USERNAME_LEN     = 20
	MAX_KEY_ID_LEN       = 5
	MAX_CONFIRMATION_LEN = 7
	MAX_OK_LEN           = 12
	MAX_RECHARGING_LEN   = 12
	MAX_FULL_POWER_LEN   = 12
	MAX_MESSAGE_LEN      = 100
)

var AUTH_KEYS = [...]map[string]int{
	{
		"server_key": 23019,
		"client_key": 32037,
	},
	{
		"server_key": 32037,
		"client_key": 29295,
	},
	{
		"server_key": 18789,
		"client_key": 13603,
	},
	{
		"server_key": 16443,
		"client_key": 29533,
	},
	{
		"server_key": 18189,
		"client_key": 21952,
	},
}

func (r *Robot) authenticate() (err error) {
	// Get robot's username
	username, err := r.getMessage(MAX_USERNAME_LEN)
	if err != nil {
		log.Printf("Error while getting robot's name: %s\n", err)
		return err
	}

	err = checkName(username)
	if err != nil {
		log.Printf("Authentication failed - wrong username '%s'\n", username)
		return err
	}

	// Set username so we can use it later in other functions as well
	r.Username = username

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

	log.Printf("[%s] Looking for key index %s\n", username, recKeyIndexStr)
	serverKey, clientKey, err := authkeyLookup(recKeyIndexStr)
	if err != nil {
		return err
	}
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
		log.Printf("[%s] Client hash is not a number: '%s'\n", username, recClientHash)
		// return err
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	log.Printf("[%s] Recieved client hash '%s'.\n", username, recClientHash)
	log.Printf("[%s] Checking if client hashes match ('%d' == '%s')\n", username, clientHash, recClientHash)
	if recClientHashInt == clientHash {
		log.Printf("[%s] Successfully authenticated.\n", username)
		_, err = r.Conn.Write([]byte(SERVER_OK))
		if err != nil {
			return err
		}
	} else {
		log.Printf("[%s] Failed to authenticate.\n", username)
		return errors.New(SERVER_LOGIN_FAILED)
	}
	return nil
}

// Checks if the name param complies with our rules
func checkName(name string) (err error) {
	if len(name) > (MAX_USERNAME_LEN - 2) {
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	return nil
}

// Looks up an auth key by the index string specified.
func authkeyLookup(iStr string) (serverKey, clientKey int, err error) {
	i, err := strconv.Atoi(iStr)
	if err != nil {
		return -1, -1, errors.New(SERVER_SYNTAX_ERROR)
	}
	if i < 0 || i > (len(AUTH_KEYS)-1) {
		return -1, -1, errors.New(SERVER_KEY_OUT_OF_RANGE_ERROR)
	}
	keys := AUTH_KEYS[i]
	serverKey = keys["server_key"]
	clientKey = keys["client_key"]
	return
}

// Calculates a hash for the username passed in.
func getHash(username string) (hash int) {
	// log.Printf("[%s] Getting hash", username)
	asciiSum := 0
	for _, r := range username {
		asciiSum += int(r)
	}
	// log.Printf("[%s] asciiSum is: '%d'", username, asciiSum)
	hash = (asciiSum * 1000) % 65536
	// log.Printf("[%s] hash is: '%d'", username, hash)
	return
}

type Direction int

const (
	UP Direction = iota
	DOWN
	LEFT
	RIGHT
)

// X, Y location
type Coordinate struct {
	x int
	y int
}

// Checks if robot moved from their previous position
func (r *Robot) moved() bool {
	return r.coors.x != r.prevCoors.x || r.coors.y != r.prevCoors.y
}

// Sets initial coordinates
func (r *Robot) setInitCoordinates() (err error) {
	moveCount := 0
	log.Printf("[%s] Getting initial coordinates...\n", r.Username)
	for moveCount < 2 {
		_, err = r.Conn.Write([]byte(SERVER_MOVE))
		if err != nil {
			return err
		}

		msg, err := r.getMessage(MAX_OK_LEN)
		if err != nil {
			log.Printf("[%s] Error while getting initial coordinates: %s\n", r.Username, err)
			return err
		}
		// log.Printf("[%s] Received a message: '%s'", r.Username, msg)

		err = r.parseAndSetCoordinates(msg)
		if err != nil {
			return err
		}
		// if r.robotMoved() {
		// 	moveCount += 1
		// }
		moveCount = moveCount + 1
	}
	log.Printf("[%s] Initial coordinates: %+v -> %+v and direction '%d'", r.Username, *(r.prevCoors), *(r.coors), r.Direction)
	return nil
}

func (r *Robot) setDirection() {
	if r.coors.x > r.prevCoors.x {
		r.Direction = RIGHT
	} else if r.coors.x < r.prevCoors.x {
		r.Direction = LEFT
	} else if r.coors.y > r.prevCoors.y {
		r.Direction = UP
	} else if r.coors.y < r.prevCoors.y {
		r.Direction = DOWN
	}
}

func (r *Robot) changeDirection() {
	switch r.Direction {
	case RIGHT:
		r.turn(SERVER_TURN_LEFT)
		r.Direction = UP
	case UP:
		r.turn(SERVER_TURN_LEFT)
		r.Direction = LEFT
	case DOWN:
		r.turn(SERVER_TURN_LEFT)
		r.Direction = RIGHT
	case LEFT:
		r.turn(SERVER_TURN_LEFT)
		r.Direction = UP
	}
}

func (r *Robot) parseAndSetCoordinates(msg string) (err error) {
	parts := strings.Split(msg, " ")

	if len(parts) > 3 || parts[0] != "OK" {
		return errors.New(SERVER_SYNTAX_ERROR)
	}

	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	if r.coors != nil {
		r.prevCoors = r.coors
	}
	r.coors = new(Coordinate)
	r.coors.x = x
	r.coors.y = y
	if r.prevCoors != nil {
		r.setDirection()
	}
	return
}

// Moves robot one step in his current direction
func (r *Robot) move() (err error) {
	// We have the right direction so just move one in the current direction
	res, err := r.executeCommandAndWaitForResponse(SERVER_MOVE, MAX_OK_LEN)
	// _, err = r.Conn.Write([]byte(SERVER_MOVE))
	if err != nil {
		return err
	}
	if err = r.parseAndSetCoordinates(res); err != nil {
		return err
	}
	return nil
}

// Turns robot into the way specified
func (r *Robot) turn(dir string) (err error) {
	// We have the right direction so just move one in the current direction
	res, err := r.executeCommandAndWaitForResponse(dir, MAX_OK_LEN)
	// _, err = r.Conn.Write([]byte(SERVER_MOVE))
	if err != nil {
		return err
	}
	if err = r.parseAndSetCoordinates(res); err != nil {
		return err
	}
	return nil
}

func (r *Robot) moveUp() (err error) {
	switch r.Direction {
	case UP:
		// We have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return
		}
	case DOWN:
		// Turn right twice to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return
		}
	case LEFT:
		// Turn right once to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	default:
		// r.Direction == RIGHT
		// Turn left once to get the right direction
		if err = r.turn(SERVER_TURN_LEFT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	}
	r.setDirection()
	return nil
}

func (r *Robot) moveDown() (err error) {
	switch r.Direction {
	case UP:
		// Turn right twice to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case DOWN:
		// We have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case LEFT:
		// Turn left once to get the right direction
		if err = r.turn(SERVER_TURN_LEFT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	default:
		// r.Direction == RIGHT
		// Turn right once to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	}
	r.setDirection()
	return nil
}

func (r *Robot) moveLeft() (err error) {
	switch r.Direction {
	case UP:
		// Turn left once to get the right direction
		if err = r.turn(SERVER_TURN_LEFT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case DOWN:
		// Turn right to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case LEFT:
		// We have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	default:
		// r.Direction == RIGHT
		// Turn right twice to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	}
	r.setDirection()
	return nil
}

func (r *Robot) moveRight() (err error) {
	switch r.Direction {
	case UP:
		// Turn right once to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case DOWN:
		// Turn left to get the right direction
		if err = r.turn(SERVER_TURN_LEFT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	case LEFT:
		// Turn right twice to get the right direction
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		if err = r.turn(SERVER_TURN_RIGHT); err != nil {
			return err
		}
		// Now we have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	default:
		// r.Direction == RIGHT
		// We have the right direction so just move one in the current direction
		if err = r.move(); err != nil {
			return err
		}
	}
	r.setDirection()
	return nil
}

// Navigates robot towards the secret message, located at [0,0]
func (r *Robot) navigateToSecretMessage() (err error) {
	log.Printf("[%s] Currently at: %+v", r.Username, *(r.coors))
	toMoveX := 0 - r.coors.x
	toMoveY := 0 - r.coors.y

	// first move diagonally = left/ right
	for toMoveX > 0 {
		log.Printf("[%s] %+v Moving right", r.Username, *(r.coors))
		// We need to move right
		if err = r.moveRight(); err != nil {
			return err
		}
		if r.moved() {
			toMoveX = toMoveX - 1
		} else {
			r.changeDirection()
			if err = r.move(); err != nil {
				return
			}
			return r.navigateToSecretMessage()
		}
	}
	for toMoveX < 0 {
		log.Printf("[%s] %+v Moving left", r.Username, *(r.coors))
		// We need to move left
		if err = r.moveLeft(); err != nil {
			return err
		}
		if r.moved() {
			toMoveX = toMoveX + 1
		} else {
			r.changeDirection()
			if err = r.move(); err != nil {
				return
			}
			return r.navigateToSecretMessage()
		}
	}

	// next we move horizontally = up/ down
	for toMoveY > 0 {
		log.Printf("[%s] %+v Moving up", r.Username, *(r.coors))
		// We need to move down
		if err = r.moveUp(); err != nil {
			return err
		}

		if r.moved() {
			toMoveY = toMoveY - 1
		} else {
			r.changeDirection()
			if err = r.move(); err != nil {
				return
			}
			return r.navigateToSecretMessage()
		}
	}
	for toMoveY < 0 {
		log.Printf("[%s] %+v Moving down", r.Username, *(r.coors))
		// We need to move down
		if err = r.moveDown(); err != nil {
			return err
		}
		if r.moved() {
			toMoveY = toMoveY + 1
		} else {
			r.changeDirection()
			if err = r.move(); err != nil {
				return
			}
			return r.navigateToSecretMessage()
		}
	}
	return nil
}

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
		// log.Printf("[%s] Getting message from buffer: '%s'\n", r.Username, r.Buffer)
		parts := strings.SplitN(r.Buffer, "\a\b", 2)
		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			msg = parts[0]
			r.Buffer = parts[1]
			if msg == strings.Replace(CLIENT_RECHARGING, "\a\b", "", 1) {
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
			log.Printf("Maximum message (%s) length exceeded! %d > %d\n", r.Buffer, len(r.Buffer), maxLength-1)
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
		// log.Printf("[%s] [RECHARGING] Reading buffer\n", r.Username)
		err = r.readSocketBuffer(TIMEOUT_RECHARGING)
		if err != nil {
			log.Printf("[%s] [RECHARGING] Error occured during reading socket buffer: %s\n", r.Username, err)
			return
		}

		parts := strings.SplitN(r.Buffer, "\a\b", 2)
		// Wait until we get the \a\b sequence on input
		if len(parts) == 2 {
			msg := parts[0]
			r.Buffer = parts[1]
			// If we receive a message that isn't CLIENT_FULL_POWER
			if msg != strings.Replace(CLIENT_FULL_POWER, "\a\b", "", 1) {
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
		return err
	}
	if e, ok := err.(interface{ Timeout() bool }); ok && e.Timeout() {
		log.Println("Timeout error", e)
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

const (
	BUFFER_SIZE        = 1024
	TIMEOUT            = 1 * time.Second // Server i klient očekávají od protistrany odpověď po dobu tohoto intervalu.
	TIMEOUT_RECHARGING = 5 * time.Second // Časový interval, během kterého musí robot dokončit dobíjení.

	// Constatnt Server messages
	SERVER_MOVE                   = "102 MOVE\a\b"             //	Příkaz pro pohyb o jedno pole vpřed
	SERVER_TURN_LEFT              = "103 TURN LEFT\a\b"        //	Příkaz pro otočení doleva
	SERVER_TURN_RIGHT             = "104 TURN RIGHT\a\b"       //	Příkaz pro otočení doprava
	SERVER_PICK_UP                = "105 GET MESSAGE\a\b"      //	Příkaz pro vyzvednutí zprávy
	SERVER_LOGOUT                 = "106 LOGOUT\a\b"           //	Příkaz pro ukončení spojení po úspěšném vyzvednutí zprávy
	SERVER_KEY_REQUEST            = "107 KEY REQUEST\a\b"      //	Žádost serveru o Key ID pro komunikaci
	SERVER_OK                     = "200 OK\a\b"               //	Kladné potvrzení
	SERVER_LOGIN_FAILED           = "300 LOGIN FAILED\a\b"     //	Nezdařená autentizace
	SERVER_SYNTAX_ERROR           = "301 SYNTAX ERROR\a\b"     //	Chybná syntaxe zprávy
	SERVER_LOGIC_ERROR            = "302 LOGIC ERROR\a\b"      //	Zpráva odeslaná ve špatné situaci
	SERVER_KEY_OUT_OF_RANGE_ERROR = "303 KEY OUT OF RANGE\a\b" // Key ID není v očekávaném rozsahu

	// Constatnt Client Messages
	CLIENT_RECHARGING = "RECHARGING\a\b" // Robot se začal dobíjet a přestal reagovat na zprávy.
	CLIENT_FULL_POWER = "FULL POWER\a\b" // Robot doplnil energii a opět příjímá příkazy.
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Printf("[%s] Handling a new connection...\n", conn.RemoteAddr().String())

	// Initialize robot
	r := Robot{Conn: conn}

	defer func() {
		log.Printf("[%s] Closing connection...\n", r.Username)
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close listener:", err)
		}
	}()

	// Handle auth
	err := r.authenticate()
	if err != nil {
		log.Printf("[%s] Error while authenticating: %s\n", r.Username, err.Error())
		r.Conn.Write([]byte(err.Error()))
		return
	}

	// Set initial coordinates
	err = r.setInitCoordinates()
	if err != nil {
		log.Printf("[%s] Error while setting initial coordinates: %s\n", r.Username, err.Error())
		r.Conn.Write([]byte(err.Error()))
		return
	}

	err = r.navigateToSecretMessage()
	if err != nil {
		log.Printf("[%s] Error while navigating to the secret message: %s\n", r.Username, err.Error())
		r.Conn.Write([]byte(err.Error()))
		return
	}
	log.Printf("[%s] About to get secret message - currently at %+v\n", r.Username, *(r.coors))

	secretMsg, err := r.executeCommandAndWaitForResponse(SERVER_PICK_UP, MAX_MESSAGE_LEN)
	if err != nil {
		log.Printf("[%s] Error while getting the secret message: %s\n", r.Username, err.Error())
		r.Conn.Write([]byte(err.Error()))
		return
	}

	log.Printf("[%s] Received the secret message: %s\n", r.Username, secretMsg)
	r.Conn.Write([]byte(SERVER_LOGOUT))
}

func main() {
	server.StartListener()
}
