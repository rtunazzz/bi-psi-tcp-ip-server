package server

import (
	"errors"
	"fmt"
	"log"
	"strconv"
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
