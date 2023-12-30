package main

import "fmt"

type Creature struct {
	ID int
	X  int
	Y  int
	Vx int
	Vy int
}

func NewCreature() Creature {
	c := Creature{}
	fmt.Scan(&c.ID, &c.X, &c.Y, &c.Vx, &c.Vy)
	return c
}
