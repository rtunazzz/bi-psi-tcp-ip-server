package server

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

// X, Y location
type Coordinate struct {
	x int
	y int
}

// Checks if robot moved from their previous position
func (r *Robot) robotMoved() bool {
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

		msg, err := r.getMessage(12)
		if err != nil {
			log.Printf("[%s] Error while getting initial coordinates: %s\n", r.Username, err)
			return err
		}
		// log.Printf("[%s] Received a message: '%s'", r.Username, msg)

		// EXPECTING MESSAGE IN A FORMAT OF 'OK X Y'
		parts := strings.Split(msg, " ")

		// log.Printf("[%s] After splitting: '%s'", r.Username, parts)
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
			r.coors = new(Coordinate)
			r.coors.x = x
			r.coors.y = y
			if r.robotMoved() {
				moveCount += 1
			}
		} else {
			r.coors = new(Coordinate)
			r.coors.x = x
			r.coors.y = y
			moveCount += 1
		}
	}

	log.Printf("[%s] Initial coordinates: %+v -> %+v", r.Username, *(r.prevCoors), *(r.coors))
	return nil
}
