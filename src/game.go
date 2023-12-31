package main

import (
	"fmt"
)

type GameState struct {
	CreatureCount int            `json:"creatureCount"`
	Creatures     []GameCreature `json:"creatures"`

	Resurface   map[int]Point
	DroneTarget map[int]int

	CreaturesTouched map[int]map[int]struct{}

	Tick   int
	States map[int]*State
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

func (g *GameState) LoadState() *State {
	defer g.NewTick()

	s := NewState()
	g.States[g.Tick] = s
	return s
}

func (g *GameState) AddResurface(id int, p Point) {
	g.Resurface[id] = p
}

func (g *GameState) RemoveResurface(id int) {
	delete(g.Resurface, id)
}

func (g *GameState) AddDroneTarget(droneID int, captureID int) {
	g.DroneTarget[droneID] = captureID
}

func (g *GameState) RemoveDroneTarget(droneID int) {
	delete(g.DroneTarget, droneID)
}

func (g *GameState) TouchCreature(droneID int, creatureID int) {
	v, ok := g.CreaturesTouched[droneID]
	if !ok {
		g.CreaturesTouched[droneID] = map[int]struct{}{creatureID: struct{}{}}
		return
	}
	v[creatureID] = struct{}{}
	g.CreaturesTouched[droneID] = v
}

func NewGame() *GameState {
	return &GameState{
		Creatures:        make([]GameCreature, 0),
		States:           make(map[int]*State),
		Resurface:        make(map[int]Point),
		DroneTarget:      make(map[int]int),
		CreaturesTouched: make(map[int]map[int]struct{}, 0),
	}
}
