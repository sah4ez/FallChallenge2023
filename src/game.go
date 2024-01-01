package main

import (
	"fmt"
	"os"
)

type GameState struct {
	CreatureCount int
	Creatures     []*GameCreature
	MapCreatures  map[int]*GameCreature

	Resurface       map[int]Point
	DroneTarget     map[int]int
	TargetCreatures map[int]struct{}

	CreaturesTouched map[int]map[int]struct{}

	MapDroneLigthCount map[string]map[int]int

	Tick   int
	States map[int]*State
}

func (g *GameState) LoadCreatures() {

	fmt.Scan(&g.CreatureCount)
	for i := 0; i < g.CreatureCount; i++ {
		gc := NewGameCreature()
		g.Creatures = append(g.Creatures, gc)
		g.MapCreatures[gc.ID] = gc
	}
}

func (g *GameState) NewTick() {
	g.Tick = g.Tick + 1
}

func (g *GameState) LoadState() *State {
	defer g.NewTick()

	fmt.Fprintln(os.Stderr, "tick: ", g.Tick)
	s := NewState(g)
	g.States[g.Tick] = s
	return s
}

func (g *GameState) AddResurface(id int, p Point) {
	g.Resurface[id] = p
}

func (g *GameState) RemoveResurface(id int) {
	delete(g.Resurface, id)
}

func (g *GameState) AddDroneTarget(droneID int, captureID int) (ok bool) {
	if _, ok = g.TargetCreatures[captureID]; ok {
		return !ok
	}
	g.TargetCreatures[captureID] = struct{}{}
	g.DroneTarget[droneID] = captureID
	return true
}

func (g *GameState) RemoveDroneTarget(drone Drone) {
	delete(g.DroneTarget, drone.ID)
}

func (g *GameState) IsTouchedCreature(creatureID int) bool {
	var found bool
	for _, cs := range g.CreaturesTouched {
		_, found = cs[creatureID]
	}
	return found
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

func (g *GameState) GetCreature(creatureID int) *GameCreature {
	if c, ok := g.MapCreatures[creatureID]; ok {
		return c
	}
	return nil
}

func (g *GameState) DebugCreatures() {
	fmt.Fprintf(os.Stderr, "got creatrues:")
	for _, c := range g.Creatures {
		fmt.Fprintf(os.Stderr, "(%d %d %d)", c.ID, c.Color, c.Type)
	}
	fmt.Fprintln(os.Stderr)
}

func (g *GameState) GetCoutLights(drone *Drone) int {
	state := drone.DetectMode()
	st, ok := g.MapDroneLigthCount[state]
	if !ok {
		return 0
	}
	cnt, ok := st[drone.ID]
	if !ok {
		return 0
	}
	defer func() {
		if cnt > 0 {
			g.MapDroneLigthCount[state][drone.ID] = cnt - 1
		}
	}()
	return cnt
}

func NewGame() *GameState {
	return &GameState{
		Creatures:        make([]*GameCreature, 0),
		MapCreatures:     make(map[int]*GameCreature, 0),
		States:           make(map[int]*State),
		Resurface:        make(map[int]Point),
		DroneTarget:      make(map[int]int),
		TargetCreatures:  make(map[int]struct{}, 0),
		CreaturesTouched: make(map[int]map[int]struct{}, 0),
		MapDroneLigthCount: map[string]map[int]int{
			ModeType0: map[int]int{
				0: 0,
				1: 0,
			},
			ModeType1: map[int]int{
				0: 1,
				1: 1,
			},
			ModeType2: map[int]int{
				0: 2,
				1: 2,
			},
			ModeType3: map[int]int{
				0: 3,
				1: 3,
			},
		},
	}
}
