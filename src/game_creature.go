package main

import (
	"fmt"
)

type GameCreature struct {
	ID    int
	Color int
	Type  int
}

func NewGameCreature() GameCreature {
	c := GameCreature{}
	fmt.Scan(&c.ID, &c.Color, &c.Type)
	return c
}
