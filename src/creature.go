package main

import (
	"fmt"
)

type Creature struct {
	ID int
	X  int
	Y  int
	Vx int
	Vy int
}

func (c *Creature) Point() Point {
	return Point{X: c.X, Y: c.Y}
}

func (c *Creature) NextPoint() Point {
	return Point{X: c.X + c.Vx, Y: c.Y + c.Vy}
}

func (c *Creature) MultNextPoint(mul float64) Point {
	return Point{X: c.X + int(mul)*c.Vx, Y: c.Y + int(mul)*c.Vy}
}

func NewCreature() Creature {
	c := Creature{}
	fmt.Scan(&c.ID, &c.X, &c.Y, &c.Vx, &c.Vy)
	return c
}
