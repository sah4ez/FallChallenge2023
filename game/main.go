package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
)

const SurfaceDistance = 500.0
const DroneSurfaceDistance = 499.0
const AutoScanDistance = 799.0
const AutoScanMonsterDistance = 1099.0
const MaxAutoScanDistance = 2000.0
const ShiftByRadarDistance = 600
const MaxPosistionX = 9999
const MaxPosistionY = 9999
const MinPosistionX = 0
const MinPosistionY = 0
const LightBattary = 5
const MinRandStep = -600
const MaxRandStep = 600
const ModeType0 = "0-2500"
const ModeType1 = "2500-5000"
const ModeType2 = "5000-7500"
const ModeType3 = "7500-10000"
const AngleRadar = 90
const RadarTL = "TL"
const RadarTR = "TR"
const RadarBL = "BL"
const RadarBR = "BR"

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
	radiusPoint  []Node
	nextPoint    Point
}

func (d *Drone) IsEmergency() bool {
	return d.Emergency == 1
}

func (d *Drone) IsSurfaced() bool {
	return d.Y <= SurfaceDistance
}

func (d *Drone) Distance(x, y int) float64 {
	return math.Sqrt(math.Pow(float64(d.X-x), 2) + math.Pow(float64(d.Y-y), 2))
}

func (d *Drone) DistanceToPoint(p Point) float64 {
	return math.Sqrt(math.Pow(float64(d.X-p.X), 2) + math.Pow(float64(d.Y-p.Y), 2))
}

func (d *Drone) NeedSurface(nearestDist int) bool {
	distanceSurface := d.Distance(d.X, SurfaceDistance)
	return distanceSurface < float64(nearestDist)
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

func (d *Drone) FindNearCaptureByRadar(g *GameState, s *State) (p Point, dd float64, cID int) {
	hashCreatres := map[int]struct{}{}
	p = Point{X: d.Y, Y: d.Y}
	exists := map[int]struct{}{}
	for _, r := range s.Radar {
		exists[r.CreatureID] = struct{}{}
	}

	monsters := []Point{}
	for _, v := range s.Creatures {
		c := g.GetCreature(v.ID)
		dist := d.Distance(v.X, v.Y)
		if c.Type < 0 && dist < AutoScanMonsterDistance {
			monsters = append(monsters, Point{X: v.X, Y: v.Y})
			fmt.Fprintf(os.Stderr, "visible near %d %d %d\n", d.ID, c.ID, c.Type)
			// p.X = d.X + v.Vx*10
			// p.Y = d.Y + v.Vy*10
		}
	}
	if len(monsters) > 0 {
		vectors := []Vector{}
		for _, m := range monsters {
			vectors = append(vectors, *NewVector(d.Position(), m))
		}
		v := Vector{}
		for _, vv := range vectors {
			v.X = v.X + vv.X
			v.Y = v.Y + vv.Y
		}
		v.X = int(v.X / len(vectors))
		v.Y = int(v.Y / len(vectors))
		p = Point{X: d.X - v.X, Y: d.Y - v.Y}
		return p, d.Distance(p.Y, p.Y), 0
	}

	for _, n := range d.radiusPoint {
		for _, c := range n.CreaturesTypes {
			if _, ok := exists[c.ID]; !ok {
				continue
			}
			if _, ok := hashCreatres[c.ID]; ok {
				// skip already touched
				continue
			}
			if _, ok := s.MyCreatures[c.ID]; ok {
				// skip scanned creature
				continue
			}
			var found bool
			for _, v := range s.DroneScnas {
				for k := range v {
					if c.ID == k {
						hashCreatres[k] = struct{}{}
						found = true
					}
				}
			}
			if found {
				continue
			}
			hashCreatres[c.ID] = struct{}{}
			dx, dy := RadarToDirection(n.Radar)
			p.X = p.X + dx*n.X
			p.Y = p.Y + dy*n.Y
		}
	}

	return p, d.Distance(p.Y, p.Y), 0
}

func (d *Drone) Position() Point {
	return Point{X: d.X, Y: d.Y}
}

func (d *Drone) TurnLight(g *GameState) {
	cnt := g.GetCoutLights(d)

	if d.Battery > LightBattary && cnt > 0 {
		d.enabledLight = true
	}
	fmt.Fprintf(os.Stderr, "turn light: %d %d %d %v\n", d.ID, d.Battery, cnt, d.enabledLight)
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

func (d *Drone) MoveToRadarByMonster(radar string, msg ...string) {
	p := ShiftByRadarByMonster(radar, d.X, d.Y)
	d.Move(p, msg...)
}

func (d *Drone) RandMove() {
	x := rand.Intn(MaxRandStep-MinRandStep) + MinRandStep
	y := rand.Intn(MaxRandStep-MinRandStep) + MinRandStep
	d.Move(Point{X: d.X + x, Y: d.Y + y}, "random")
}

func (d *Drone) GetRadiusLight() float64 {
	radius := AutoScanDistance
	if d.enabledLight {
		radius = MaxAutoScanDistance
	}
	return radius
}

func (d *Drone) SolveRadarRadius(g *GameState, radar []Radar) {
	hashRadar := map[string][]*GameCreature{}
	for _, r := range radar {
		if v, ok := hashRadar[r.Radar]; !ok {
			c := g.GetCreature(r.CreatureID)
			if c == nil {
				continue
			}
			hashRadar[r.Radar] = []*GameCreature{c}
		} else {
			c := g.GetCreature(r.CreatureID)
			if c == nil {
				continue
			}
			v = append(v, c)
			hashRadar[r.Radar] = v
		}
	}

	radius := d.GetRadiusLight()
	for tetha := 45; tetha < 360; tetha += AngleRadar {
		radarName := AngelToRadar(tetha)

		xt := math.Cos(float64(tetha)) * radius
		yt := math.Sin(float64(tetha)) * radius
		ct := []*GameCreature{}
		ct = append(ct, hashRadar[radarName]...)
		d.radiusPoint = append(d.radiusPoint, Node{
			Point: Point{
				X: d.X + int(xt),
				Y: d.Y + int(yt),
			},
			CreaturesTypes: ct,
			Radar:          radarName,
		})
	}
}

func (d *Drone) DebugRadarRadius() {
	radius := d.GetRadiusLight()
	fmt.Fprintf(os.Stderr, ">1%d(%d,%d):", d.ID, d.X, d.Y)
	for _, p := range d.radiusPoint {
		types := []int{}
		for _, g := range p.CreaturesTypes {
			types = append(types, g.Type)
		}
		fmt.Fprintf(os.Stderr, "(%s,%d,%d,%v),", p.Radar, p.X, p.Y, types)
	}
	fmt.Fprintf(os.Stderr, "%f\n", radius)
}

func (d *Drone) Move(p Point, msg ...string) {
	if len(msg) == 0 {
		msg = append(msg, fmt.Sprintf("p: %d %d to: %d %d", d.X, d.Y, p.X, p.Y))
		fmt.Fprintf(os.Stderr, strings.Join(msg, " ")+"\n")
	}
	if p.X < 0 {
		p.X = 0
	}
	if p.X > MaxPosistionX-AutoScanDistance+3 {
		p.X = MaxPosistionX - AutoScanDistance + 3
	}

	if p.Y < 0 {
		p.Y = 0
	}
	if p.Y > MaxPosistionY-AutoScanDistance+3 {
		p.Y = MaxPosistionY - AutoScanDistance + 3
	}
	fmt.Printf("MOVE %d %d %s\n", p.X, p.Y, d.Light())
}

func (d *Drone) DetectMode() string {
	if d.Y < 2500 {
		return ModeType0
	} else if d.Y < 5000 {
		return ModeType1
	} else if d.Y < 7500 {
		return ModeType2
	} else {
		return ModeType3
	}
}

func (d *Drone) Debug() {

}

func NewDrone() Drone {
	d := Drone{
		radiusPoint: make([]Node, 0),
	}
	fmt.Scan(&d.ID, &d.X, &d.Y, &d.Emergency, &d.Battery)
	// fmt.Fprintln(os.Stderr, "scan drone", d.ID, d.X, d.Y, d.Emergency, d.Battery)
	return d
}

type GameState struct {
	CreatureCount int
	Creatures     []*GameCreature
	MapCreatures  map[int]*GameCreature

	Resurface       map[int]Point
	DroneTarget     map[int]int
	TargetCreatures map[int]struct{}

	CreaturesTouched map[int]map[int]struct{}

	MapDroneLigthCount map[string]map[int]int

	DroneQueue map[int][]Point

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

func (g *State) GetCreaturePosition(creatureID int) *Creature {
	for _, v := range g.Creatures {
		if v.ID == creatureID {
			return &v
		}
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
		TargetCreatures:  make(map[int]struct{}, 0),
		CreaturesTouched: make(map[int]map[int]struct{}, 0),
		MapDroneLigthCount: map[string]map[int]int{
			ModeType0: map[int]int{},
			ModeType1: map[int]int{},
			ModeType2: map[int]int{},
			ModeType3: map[int]int{},
		},
		DroneQueue: make(map[int][]Point, 0),
	}
}

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

func main() {
	game := NewGame()

	game.LoadCreatures()
	game.DebugCreatures()

	var once sync.Once
	var onceDrone sync.Once

	for {
		s := game.LoadState()
		// s.DebugRadar()
		s.DebugVisibleCreatures()

		once.Do(game.DebugLightCoutns)
		onceDrone.Do(func() {
			for i := range s.MyDrones {
				drone := s.MyDrones[i]
				game.MoveDrone(drone, Point{X: drone.X, Y: 2500})
				game.MoveDrone(drone, Point{X: drone.X, Y: 8500})
				game.MoveDrone(drone, Point{X: int(MaxPosistionX / 2), Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: int(MaxPosistionX / 2), Y: SurfaceDistance})
			}
		})

		for i := range s.MyDrones {
			drone := s.MyDrones[i]
			if drone.IsEmergency() {
				game.RemoveDroneTarget(drone)
				fmt.Fprintln(os.Stderr, drone.ID, "emergency", s.DroneScnas[drone.ID])
				drone.Wait()
				continue
			}

			drone.TurnLight(game)
			drone.SolveRadarRadius(game, s.MapRadar[drone.ID])
			drone.DebugRadarRadius()

			newPoint := game.FirstCommand(drone)
			if drone.DistanceToPoint(newPoint) < AutoScanDistance {
				drone.TurnLight(game)
				newPoint = game.PopCommand(drone)
			}
			drone.TurnLight(game)
			drone.Move(newPoint)
		}
	}
}

type Node struct {
	Point

	CreaturesTypes []*GameCreature
	Score          int
	Radar          string

	Drone    *Drone
	FoeDrone *Drone
	Creature *Creature
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

func ShiftByRadarByMonster(radar string, x, y int) (p Point) {

	p = Point{X: x, Y: y}
	switch radar {
	case "TR":
		p.X, p.Y = p.X-ShiftByRadarDistance, p.Y+ShiftByRadarDistance
	case "TL":
		p.X, p.Y = p.X+ShiftByRadarDistance, p.Y+ShiftByRadarDistance
	case "BR":
		p.X, p.Y = p.X-ShiftByRadarDistance, p.Y-ShiftByRadarDistance
	case "BL":
		p.X, p.Y = p.X+ShiftByRadarDistance, p.Y-ShiftByRadarDistance
	}
	fmt.Fprintln(os.Stderr, "by monster radar", radar, p.String(), (Point{X: x, Y: y}).String())
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

func AngelToRadar(angle int) string {
	if angle <= 90 {
		return RadarTR
	} else if angle <= 180 {
		return RadarTL
	} else if angle <= 270 {
		return RadarBL
	} else {
		return RadarBR
	}
}

func RadarToDirection(radar string) (x int, y int) {
	switch radar {
	case RadarTL:
		return -1, -1
	case RadarTR:
		return 1, -1
	case RadarBR:
		return 1, 1
	default:
		return -1, 1
	}
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
	MapRadar       map[int][]Radar
}

func (s *State) DebugRadar() {
	fmt.Fprintf(os.Stderr, "got radar: ")
	for _, r := range s.Radar {
		fmt.Fprintf(os.Stderr, "(%s %d %d),", r.Radar, r.DroneID, r.CreatureID)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) DebugRadarByDroneID(droneID int) {
	var d Drone
	for _, d = range s.MyDrones {
		if d.ID == droneID {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "got radar %d (%s): ", droneID, d.DetectMode())
	for _, r := range s.MapRadar[droneID] {
		fmt.Fprintf(os.Stderr, "(%s %d),", r.Radar, r.CreatureID)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) DebugVisibleCreatures() {
	fmt.Fprintf(os.Stderr, "Visible: ")
	for _, c := range s.Creatures {
		fmt.Fprintf(os.Stderr, "(%d %d %d %d %d),", c.ID, c.X, c.Y, c.Vx, c.Vy)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) CheckCreatureID(droneID int, creatureID int) bool {
	v, ok := s.MapRadar[droneID]
	if !ok {
		return ok
	}
	for _, r := range v {
		if r.CreatureID == creatureID {
			return true
		}
	}
	return false
}

func NewState(g *GameState) *State {
	s := &State{
		MyCreatures:  make(map[int]struct{}, 0),
		FoeCreatures: make([]int, 0),
		MyDrones:     make([]Drone, 0),
		FoeDrones:    make([]Drone, 0),
		DroneScnas:   make(map[int]map[int]struct{}, 0),
		Creatures:    make([]Creature, 0),
		Radar:        make([]Radar, 0),
		MapRadar:     make(map[int][]Radar, 0),
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
		d := NewDrone()
		s.MyDrones = append(s.MyDrones, d)
		g.AddDroneCounts(d.ID)
	}

	fmt.Scan(&s.FoeDroneCount)
	for i := 0; i < s.FoeDroneCount; i++ {
		d := NewDrone()
		s.FoeDrones = append(s.FoeDrones, d)
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
		c := NewCreature()
		s.Creatures = append(s.Creatures, c)
	}
	sort.SliceStable(s.Creatures, func(i, j int) bool {
		ir := g.MapCreatures[s.Creatures[i].ID]
		jr := g.MapCreatures[s.Creatures[j].ID]
		return ir.Type < jr.Type
	})

	fmt.Scan(&s.RadarBlipCount)

	for i := 0; i < s.RadarBlipCount; i++ {
		r := NewRadar()
		s.Radar = append(s.Radar, r)
		if v, ok := s.MapRadar[r.DroneID]; !ok {
			s.MapRadar[r.DroneID] = []Radar{r}
		} else {
			v = append(v, r)
			s.MapRadar[r.DroneID] = v
		}
	}
	sort.SliceStable(s.Radar, func(i, j int) bool {
		ir := g.MapCreatures[s.Radar[i].CreatureID]
		jr := g.MapCreatures[s.Radar[j].CreatureID]
		return ir.Type < jr.Type
	})

	return s
}

type Vector struct {
	X int
	Y int
}

func NewVector(from Point, to Point) *Vector {
	return &Vector{
		X: to.X - from.X,
		Y: to.Y - from.Y,
	}
}

func (v *Vector) Add(to Vector) Vector {
	return Vector{
		X: v.X + to.X,
		Y: v.Y + to.Y,
	}
}
