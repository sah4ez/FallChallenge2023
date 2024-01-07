package main

import (
	"fmt"
	"os"
)

type GameState struct {
	CreatureCount   int
	Creatures       []*GameCreature
	MapCreatures    map[int]*GameCreature
	PrevMonster     []Creature
	PrevPrevMonster []Creature

	Resurface       map[int]Point
	DroneTarget     map[int]int
	DroneTarget2    map[int][]int
	TargetCreatures map[int]struct{}

	CreaturesTouched map[int]map[int]struct{}

	MapDroneLigthCount map[string]map[int]int

	DroneQueue         map[int][]Point
	DorneNextLightTick map[int]int

	Tick   int
	States map[int]*State
}

func (g *GameState) CoundScoringCreature() int {
	return 6
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

func (g *State) GetCreaturePosition(creatureID int) *Creature {
	for _, v := range g.Creatures {
		if v.ID == creatureID {
			return &v
		}
	}
	return nil
}

func (g *State) GetDrone(dorneID int) *Drone {
	for _, v := range g.MyDrones {
		if v.ID == dorneID {
			return &v
		}
	}
	return nil
}

func (g *GameState) OnDroneDepth(creature *GameCreature, drone *Drone) bool {
	switch creature.Type {
	case -1:
		return false
	case 0:
		return 2500 <= drone.Y && drone.Y <= 5000
	case 1:
		return 5000 <= drone.Y && drone.Y <= 7500
	case 2:
		return 7500 <= drone.Y && drone.Y <= 10000
	default:
		return false
	}
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
		fmt.Fprintln(os.Stderr, "count lignts 1", drone.ID, state)
		return 0
	}
	cnt, ok := st[drone.ID]
	if !ok {
		fmt.Fprintln(os.Stderr, "count lignts 2", drone.ID, state)
		return 0
	}
	defer func() {
		if cnt > 0 {
			g.MapDroneLigthCount[state][drone.ID] = cnt - 1
		}
	}()
	return cnt
}

func (g *GameState) AddDroneCounts(droneID int) {
	if v, ok := g.MapDroneLigthCount[ModeType0]; ok {
		if _, ok := v[droneID]; !ok {
			v[droneID] = 0
		}
		g.MapDroneLigthCount[ModeType0] = v
	}
	if v, ok := g.MapDroneLigthCount[ModeType1]; ok {
		if _, ok := v[droneID]; !ok {
			v[droneID] = 3
		}
		g.MapDroneLigthCount[ModeType1] = v
	}
	if v, ok := g.MapDroneLigthCount[ModeType2]; ok {
		if _, ok := v[droneID]; !ok {
			v[droneID] = 6
		}
		g.MapDroneLigthCount[ModeType2] = v
	}
	if v, ok := g.MapDroneLigthCount[ModeType3]; ok {
		if _, ok := v[droneID]; !ok {
			v[droneID] = 9
		}
		g.MapDroneLigthCount[ModeType3] = v
	}
}

func (g *GameState) DebugLightCoutns() {
	fmt.Fprintf(os.Stderr, "init light counts")
	for k, v := range g.MapDroneLigthCount {
		fmt.Fprintf(os.Stderr, "%s: ", k)
		for kk, vv := range v {
			fmt.Fprintf(os.Stderr, "%d %d|", kk, vv)
		}
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintln(os.Stderr)
}

func (g *GameState) MoveDrone(drone Drone, p Point) {
	if v, ok := g.DroneQueue[drone.ID]; !ok {
		g.DroneQueue[drone.ID] = []Point{p}
		return
	} else {
		v = append(v, p)
		g.DroneQueue[drone.ID] = v
	}
}

func (g *GameState) FirstCommand(drone Drone) (p Point) {
	v, ok := g.DroneQueue[drone.ID]
	if !ok {
		return
	}
	if len(v) == 0 {
		return
	}
	return v[0]
}

func (g *GameState) PopCommand(drone Drone) (p Point) {
	v, ok := g.DroneQueue[drone.ID]
	if !ok {
		return
	}
	if len(v) <= 1 {
		g.DroneQueue[drone.ID] = v[:0]
		return
	}
	g.DroneQueue[drone.ID] = v[1:]
	return v[0]
}

func NewGame() *GameState {
	return &GameState{
		Creatures:        make([]*GameCreature, 0),
		MapCreatures:     make(map[int]*GameCreature, 0),
		States:           make(map[int]*State),
		Resurface:        make(map[int]Point),
		DroneTarget:      make(map[int]int),
		DroneTarget2:     make(map[int][]int),
		TargetCreatures:  make(map[int]struct{}, 0),
		CreaturesTouched: make(map[int]map[int]struct{}, 0),
		MapDroneLigthCount: map[string]map[int]int{
			ModeType0: map[int]int{},
			ModeType1: map[int]int{},
			ModeType2: map[int]int{},
			ModeType3: map[int]int{},
		},
		DroneQueue:         make(map[int][]Point, 0),
		PrevMonster:        make([]Creature, 0),
		PrevPrevMonster:    make([]Creature, 0),
		DorneNextLightTick: make(map[int]int, 0),
	}
}
