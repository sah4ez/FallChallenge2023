package main

import (
	"fmt"
	"strings"
)

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

// MOVE <x> <y> <light (1|0)> | WAIT <light (1|0)>
type Drone struct {
	ID        int
	X         int
	Y         int
	Emergency int
	Battery   int

	enabledLight bool
}

func (d *Drone) TurnLight() {
	if d.enabledLight {
		d.enabledLight = false
		return
	}
	d.enabledLight = true
}

func (d *Drone) Light() string {

	if d.enabledLight {
		return "1"
	}
	return "0"
}

func (d *Drone) Wait(msg ...string) {
	if len(msg) == 0 {
		msg = append(msg, fmt.Sprintf("%d %d", d.X, d.Y))
	}
	fmt.Printf("WAIT %s %s\n", d.Light(), strings.Join(msg, " "))
}

func (d *Drone) Move(x, y int, msg ...string) {
	if len(msg) == 0 {
		msg = append(msg, fmt.Sprintf("%d %d", d.X, d.Y))
	}
	fmt.Printf("MOVE %d %d %s %s\n", x, y, d.Light(), strings.Join(msg, " "))
}

func (d *Drone) Debug() {

}

func NewDrone() Drone {
	d := Drone{}
	fmt.Scan(&d.ID, &d.X, &d.Y, &d.Emergency, &d.Battery)
	return d
}

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

func main() {
	game := NewGame()

	game.LoadCreatures()

	for {
		s := game.LoadState()

		for _, d := range s.MyDrones {
			d.TurnLight()
			d.Wait()
		}
	}
}

type Radar struct {
	DroneID    int
	CreatureID int
	Radar      string
}

func NewRadar() Radar {
	r := Radar{}
	fmt.Scan(&r.DroneID, &r.CreatureID, &r.Radar)
	return r
}

type State struct {
	MyScore  int
	FoeScore int

	MyScanCount int
	MyCreatures []int

	FoeScanCount int
	FoeCreatures []int

	MyDroneCount int
	MyDrones     []Drone

	FoeDroneCount int
	FoeDrones     []Drone

	DroneScanCount int
	DroneScnas     map[int]int

	VisibleCreatureCount int
	Creatures            []Creature

	RadarBlipCount int
	Radar          []Radar
}

func NewState() State {
	s := State{
		MyCreatures:  make([]int, 0),
		FoeCreatures: make([]int, 0),
		MyDrones:     make([]Drone, 0),
		FoeDrones:    make([]Drone, 0),
		DroneScnas:   make(map[int]int, 0),
		Creatures:    make([]Creature, 0),
		Radar:        make([]Radar, 0),
	}

	fmt.Scan(&s.MyScore)
	fmt.Scan(&s.FoeScore)

	fmt.Scan(&s.MyScanCount)
	for i := 0; i < s.MyScanCount; i++ {
		var creatureId int
		fmt.Scan(&creatureId)
		s.MyCreatures = append(s.MyCreatures, creatureId)
	}

	fmt.Scan(&s.FoeScanCount)
	for i := 0; i < s.FoeScanCount; i++ {
		var creatureId int
		fmt.Scan(&creatureId)
		s.FoeCreatures = append(s.FoeCreatures, creatureId)
	}

	fmt.Scan(&s.MyDroneCount)
	for i := 0; i < s.MyDroneCount; i++ {
		s.MyDrones = append(s.MyDrones, NewDrone())
	}

	fmt.Scan(&s.FoeDroneCount)
	for i := 0; i < s.FoeDroneCount; i++ {
		s.FoeDrones = append(s.FoeDrones, NewDrone())
	}

	fmt.Scan(&s.DroneScanCount)
	for i := 0; i < s.DroneScanCount; i++ {
		var droneId, creatureId int
		fmt.Scan(&droneId, &creatureId)
		s.DroneScnas[droneId] = creatureId
	}

	fmt.Scan(&s.VisibleCreatureCount)

	for i := 0; i < s.VisibleCreatureCount; i++ {
		s.Creatures = append(s.Creatures, NewCreature())
	}

	fmt.Scan(&s.RadarBlipCount)

	for i := 0; i < s.RadarBlipCount; i++ {
		s.Radar = append(s.Radar, NewRadar())
	}

	return s
}
