package main

import (
	"fmt"
)

type GameCreature struct {
	ID    int
	Color int
	Type  int
}

func (c *GameCreature) IsMonster() bool {
	if c == nil {
		return false
	}
	return c.Type < 0
}

func NewGameCreature() *GameCreature {
	c := &GameCreature{}
	fmt.Scan(&c.ID, &c.Color, &c.Type)
	return c
}
