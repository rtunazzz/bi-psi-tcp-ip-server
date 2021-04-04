package server

import (
	"errors"
	"log"
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

var AUTH_KEYS = map[string]map[string]int{
	"0": {
		"server_key": 23019,
		"client_key": 32037,
	},
	"1": {
		"server_key": 32037,
		"client_key": 29295,
	},
	"2": {
		"server_key": 18789,
		"client_key": 13603,
	},
	"3": {
		"server_key": 16443,
		"client_key": 29533,
	},
	"4": {
		"server_key": 18189,
		"client_key": 21952,
	},
}

// Checks if the name param complies with our rules
func checkName(name string) (err error) {
	if len(name) > (MAX_USERNAME_LEN - 2) {
		return errors.New(SERVER_SYNTAX_ERROR)
	}
	return nil
}

func authkeyLookup(iStr string) (serverKey, clientKey int) {
	// keys map[string]int
	keys := AUTH_KEYS[iStr]
	serverKey = keys["server_key"]
	clientKey = keys["client_key"]
	return
}

func getHash(username string) (hash int) {
	log.Printf("[%s] Getting hash", username)
	asciiSum := 0
	for _, r := range username {
		asciiSum += int(r)
	}
	log.Printf("[%s] asciiSum is: '%d'", username, asciiSum)
	hash = (asciiSum * 1000) % 65536
	log.Printf("[%s] hash is: '%d'", username, hash)
	return
}
