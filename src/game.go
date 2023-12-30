package main

import (
	"fmt"
)

type GameState struct {
	CreatureCount int            `json:"creatureCount"`
	Creatures     []GameCreature `json:"creatures"`

	Tick   int
	States map[int]State
}

func (g *GameState) LoadCreatures() {

	fmt.Scan(&g.CreatureCount)
	for i := 0; i < g.CreatureCount; i++ {
		g.Creatures = append(g.Creatures, NewGameCreature())
	}
}

func (g *GameState) NewTick() {
	g.Tick = g.Tick + 1
}

func (g *GameState) LoadState() State {
	defer g.NewTick()

	s := NewState()
	g.States[g.Tick] = s
	return s
}

func NewGame() *GameState {
	return &GameState{
		Creatures: make([]GameCreature, 0),
		States:    make(map[int]State),
	}
}
