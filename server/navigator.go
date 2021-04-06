package server

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

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
