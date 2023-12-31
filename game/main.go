package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

const SurfaceDistance = 500.0
const DroneSurfaceDistance = 499.0
const AutoScanDistance = 799.0
const MaxAutoScanDistance = 2000.0
const ShiftByRadarDistance = 600
const MaxPosistionX = 9999
const MaxPosistionY = 9999
const MinPosistionX = 0
const MinPosistionY = 0
const LightBattary = 5

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
	scanned      map[int]struct{}
}

func (d *Drone) IsSurfaced() bool {
	return d.Y <= SurfaceDistance
}

func (d *Drone) Distance(x, y int) float64 {
	return math.Sqrt(math.Pow(float64(d.X-x), 2) + math.Pow(float64(d.Y-y), 2))
}

func (d *Drone) FindNearCapture(g *GameState, s *State) (p Point, dd float64, cID int) {
	targetCaptureID, ok := g.DroneTarget[d.ID]
	if ok {
		for _, c := range s.Creatures {
			if c.ID != targetCaptureID {
				continue
			}
			dd = d.Distance(c.X, c.Y)
			p.X, p.Y, cID = c.X, c.Y, c.ID
		}
		return
	}

	min := math.MaxFloat64
	for _, c := range s.Creatures {
		if _, ok := g.TargetCreatures[c.ID]; ok {
			fmt.Fprintln(os.Stderr, c.ID, "skip as target in other drone")
			continue
		}
		if v, ok := g.CreaturesTouched[d.ID]; ok {
			if len(v) > 0 {
				if _, ok := v[c.ID]; ok {
					fmt.Fprintln(os.Stderr, c.ID, "skip as touched")
					continue
				}
			}
		}
		if _, ok := s.MyCreatures[c.ID]; ok {
			fmt.Fprintln(os.Stderr, c.ID, "creatures exists")
			continue
		}
		if newMin := d.Distance(c.X, c.Y); newMin < min {
			min = newMin
			p.X, p.Y, dd, cID = c.X, c.Y, newMin, c.ID
			fmt.Fprintln(os.Stderr, c.ID, "new creatures")
		}
	}
	return
}

func (d *Drone) TurnLight() {
	if d.Battery > LightBattary && d.Y > int(MaxPosistionY/2)-AutoScanDistance {
		d.enabledLight = true
	}
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
		fmt.Fprintf(os.Stderr, strings.Join(msg, " ")+"\n")
	}
	fmt.Printf("WAIT %s\n", d.Light())
}

func (d *Drone) MoveToRadar(radar string, msg ...string) {
	p := ShiftByRadar(radar, d.X, d.Y)
	d.Move(p, msg...)
}

func (d *Drone) Move(p Point, msg ...string) {
	if len(msg) == 0 {
		msg = append(msg, fmt.Sprintf("p: %d %d to: %d %d", d.X, d.Y, p.X, p.Y))
		fmt.Fprintf(os.Stderr, strings.Join(msg, " ")+"\n")
	}
	fmt.Printf("MOVE %d %d %s\n", p.X, p.Y, d.Light())
}

func (d *Drone) Debug() {

}

func NewDrone() Drone {
	d := Drone{}
	fmt.Scan(&d.ID, &d.X, &d.Y, &d.Emergency, &d.Battery)
	fmt.Fprintln(os.Stderr, "scan drone", d.ID, d.X, d.Y, d.Emergency, d.Battery)
	return d
}

type GameState struct {
	CreatureCount int            `json:"creatureCount"`
	Creatures     []GameCreature `json:"creatures"`

	Resurface       map[int]Point
	DroneTarget     map[int]int
	TargetCreatures map[int]struct{}

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

func (g *GameState) AddDroneTarget(droneID int, captureID int) (ok bool) {
	if _, ok = g.TargetCreatures[captureID]; ok {
		return !ok
	}
	g.TargetCreatures[captureID] = struct{}{}
	g.DroneTarget[droneID] = captureID
	return true
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
		TargetCreatures:  make(map[int]struct{}, 0),
		CreaturesTouched: make(map[int]map[int]struct{}, 0),
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
		// s.DebugRadar()

		for i := range s.MyDrones {
			drone := s.MyDrones[i]

			if p, ok := game.Resurface[drone.ID]; ok {
				if !drone.IsSurfaced() {
					drone.Move(p)
					fmt.Fprintln(os.Stderr, drone.ID, "surfaced", drone.X, drone.Y)
					continue
				}
				game.RemoveResurface(drone.ID)
			}

			if len(s.DroneScnas[drone.ID]) >= 2 {
				game.AddResurface(drone.ID, Point{X: drone.X, Y: int(SurfaceDistance)})
			}
			newPoint, dist, cID := drone.FindNearCapture(game, s)
			if !newPoint.IsZero() {
				targetCaptureID, ok := game.DroneTarget[drone.ID]
				if dist <= AutoScanDistance && targetCaptureID == cID && ok {
					game.RemoveDroneTarget(drone.ID)
				}
				if dist <= AutoScanDistance {
					game.TouchCreature(drone.ID, cID)
				}
				fmt.Fprintln(os.Stderr, drone.ID, "move to nearest")
				drone.TurnLight()
				drone.Move(newPoint)
			} else {
				var hasRadar bool
				s.DebugRadar()
				for _, r := range s.Radar {
					if r.DroneID != drone.ID {
						continue
					}
					if hasRadar {
						break
					}
					var found bool
					for _, cs := range game.CreaturesTouched {
						_, found = cs[r.CreatureID]
					}
					if found {
						continue
					}
					cID, ok := game.DroneTarget[drone.ID]
					if r.DroneID == drone.ID && !ok {
						if added := game.AddDroneTarget(drone.ID, r.CreatureID); !added {
							fmt.Fprintln(os.Stderr, "not added target", len(game.TargetCreatures))
							continue
						}

						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "no target")
						hasRadar = true
						continue
					} else if r.DroneID == drone.ID && ok && r.CreatureID == cID {
						drone.TurnLight()
						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "target", cID)
						hasRadar = true
						continue
					} else {
						fmt.Fprintf(os.Stderr, "%v\n", game.DroneTarget)
					}
				}
				if !hasRadar {
					drone.Wait()
				}
			}
		}
	}
}

type Point struct {
	X int
	Y int
}

func (p Point) IsZero() bool {
	return p.X == 0 && p.Y == 0
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

func ShiftByRadar(radar string, x, y int) (p Point) {

	p = Point{X: x, Y: y}
	switch radar {
	case "TR":
		p.X, p.Y = p.X+ShiftByRadarDistance, p.Y-ShiftByRadarDistance
	case "TL":
		p.X, p.Y = p.X-ShiftByRadarDistance, p.Y-ShiftByRadarDistance
	case "BR":
		p.X, p.Y = p.X+ShiftByRadarDistance, p.Y+ShiftByRadarDistance
	case "BL":
		p.X, p.Y = p.X-ShiftByRadarDistance, p.Y+ShiftByRadarDistance
	}
	fmt.Fprintln(os.Stderr, "by radar", radar, p.String(), (Point{X: x, Y: y}).String())
	if p.X <= MinPosistionX || p.X >= MaxPosistionX {
		fmt.Fprintln(os.Stderr, "out x", p.String())
		p.X = int(MaxPosistionX / 2)
	}
	if p.Y <= MinPosistionY || p.Y >= MaxPosistionY {
		fmt.Fprintln(os.Stderr, "out y", p.String())
		p.Y = int(MaxPosistionY / 2)
	}
	return p
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
	MyCreatures map[int]struct{}

	FoeScanCount int
	FoeCreatures []int

	MyDroneCount int
	MyDrones     []Drone

	FoeDroneCount int
	FoeDrones     []Drone

	DroneScanCount int
	DroneScnas     map[int]map[int]struct{}

	VisibleCreatureCount int
	Creatures            []Creature

	RadarBlipCount int
	Radar          []Radar
	MapRadar       map[int]Radar
}

func (s *State) DebugRadar() {
	for _, r := range s.Radar {
		fmt.Fprintf(os.Stderr, "%d %d %s\n", r.DroneID, r.CreatureID, r.Radar)
	}
}

func NewState() *State {
	s := &State{
		MyCreatures:  make(map[int]struct{}, 0),
		FoeCreatures: make([]int, 0),
		MyDrones:     make([]Drone, 0),
		FoeDrones:    make([]Drone, 0),
		DroneScnas:   make(map[int]map[int]struct{}, 0),
		Creatures:    make([]Creature, 0),
		Radar:        make([]Radar, 0),
		MapRadar:     make(map[int]Radar, 0),
	}

	fmt.Scan(&s.MyScore)
	fmt.Scan(&s.FoeScore)

	fmt.Scan(&s.MyScanCount)
	for i := 0; i < s.MyScanCount; i++ {
		var creatureId int
		fmt.Scan(&creatureId)
		s.MyCreatures[creatureId] = struct{}{}
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
		if _, ok := s.DroneScnas[droneId]; ok {
			s.DroneScnas[droneId][creatureId] = struct{}{}
		} else {
			s.DroneScnas[droneId] = map[int]struct{}{creatureId: struct{}{}}
		}
	}

	fmt.Scan(&s.VisibleCreatureCount)

	for i := 0; i < s.VisibleCreatureCount; i++ {
		s.Creatures = append(s.Creatures, NewCreature())
	}

	fmt.Scan(&s.RadarBlipCount)

	for i := 0; i < s.RadarBlipCount; i++ {
		r := NewRadar()
		s.Radar = append(s.Radar, r)
		s.MapRadar[r.DroneID] = r
	}

	return s
}
