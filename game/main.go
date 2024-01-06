package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
)

const SurfaceDistance = 500.0
const ResurfaceDistance = 300
const DroneSurfaceDistance = 499.0
const AutoScanDistance = 800.0
const StepScan = float64(150.0)
const AutoScanMonsterDistance = 1099.0
const MaxAutoScanDistance = 2000.0
const ShiftByRadarDistance = 600
const MonsterDistanceDetect = 950
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
const MinLigthDepth = 2200
const DroneSize = 300.0
const DroneLightTick = 3
const MaxDepthInPath = 8300

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

// MOVE <x> <y> <light (1|0)> | WAIT <light (1|0)>
type Drone struct {
	ID          int
	X           int
	Y           int
	Emergency   int
	Battery     int
	NearMonster bool

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
	if d.Y < MinLigthDepth {
		return
	}
	var enable bool
	if v, ok := g.DorneNextLightTick[d.ID]; ok {
		if v == 1 {
			delete(g.DorneNextLightTick, d.ID)
		} else {
			v -= 1
			g.DorneNextLightTick[d.ID] = v
		}
	} else {
		g.DorneNextLightTick[d.ID] = DroneLightTick
		enable = true
	}

	if d.Battery >= LightBattary && enable {
		d.enabledLight = true
	}
	fmt.Fprintf(os.Stderr, "turn light: %d %d  %v\n", d.ID, d.Battery, d.enabledLight)
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
	// if d.enabledLight {
	// radius = MaxAutoScanDistance
	// }
	return radius
}

func (d *Drone) SolveToGraph(g *GameState, s *State, location [][]*Node, target Point) *Vertex {
	i, j := 0, 0
	if len(location)%2 == 0 {
		i = len(location) / 2
	} else {
		i = len(location)/2 + 1
	}
	if len(location[i])%2 == 0 {
		j = len(location[i]) / 2
	} else {
		j = len(location[i])/2 + 1
	}

	return NewGraph(i, j, location, nil, nil)
}

func (d *Drone) SolveFillLocation(location [][]*Node, used map[Point]struct{}) []*Node {

	i, j := 0, 0
	if len(location)%2 == 0 {
		i = len(location)/2 - 1
	} else {
		i = len(location) / 2
	}
	if len(location[i])%2 == 0 {
		j = len(location[i])/2 - 1
	} else {
		j = len(location[i]) / 2
	}
	return FillLocation(i, j, location, used)
}

func (d *Drone) Solve(g *GameState, s *State, radar []Radar, target Point) [][]*Node {
	distance := int(AutoScanDistance)
	// if d.enabledLight {
	// distance = int(MaxAutoScanDistance)
	// }
	startX := d.X - distance
	if startX < 0 {
		startX = 0
	}
	startY := d.Y - distance
	if startY < 0 {
		startY = 0
	}
	endX := d.X + distance
	if endX >= MaxPosistionX {
		endX = MaxPosistionX
	}
	endY := d.Y + distance
	if endY >= MaxPosistionY {
		endY = MaxPosistionY
	}

	sizeX := int(math.Floor(float64(endX-startX) / StepScan))
	if sizeX%2 == 0 {
		sizeX += 1
		endX += int(StepScan / 2)
		startX -= int(StepScan / 2)
	}
	sizeY := int(math.Floor(float64(endY-startY) / StepScan))
	if sizeY%2 == 0 {
		sizeY += 1
		endY += int(StepScan / 2)
		startY -= int(StepScan / 2)
	}
	location := make([][]*Node, sizeX)
	i := 0
	monsters := []Creature{}
	for _, c := range s.Creatures {
		gc := g.GetCreature(c.ID)
		if gc.Type < 0 && d.DistanceToPoint(c.Point()) < MaxAutoScanDistance {
			monsters = append(monsters, c)
		}
	}
	fmt.Fprintln(os.Stderr, "monsters", len(monsters), "prev", len(g.PrevMonster), "prevprev", len(g.PrevPrevMonster))
	if len(monsters) == 0 {
		prevStep := 2.0
		for _, m := range g.PrevMonster {
			nextMonsterPoint := m.MultNextPoint(prevStep)
			fmt.Fprintf(os.Stderr, "d:%d,m:%d (%d:%d)->(%d:%d)%f\n", d.ID, m.ID, m.X, m.Y, nextMonsterPoint.X, nextMonsterPoint.Y, d.DistanceToPoint(nextMonsterPoint))
		}
	}
	for x := startX; x < endX; x += int(StepScan) {
		if i >= len(location) {
			continue
		}
		location[i] = make([]*Node, sizeY)
		j := 0
		for y := startY; y < endY; y += int(StepScan) {
			if j >= len(location[i]) {
				continue
			}
			from := Point{X: x, Y: y}
			score := 0
			if !d.NearMonster {
				if v, ok := g.DroneTarget[d.ID]; ok {
					if radar, ok := s.DroneCreatureRadar[d.ID][v]; ok {
						score += RadarToScore(radar, from, d)
					}
				}
			}
			for _, foeDrone := range s.FoeDrones {
				if d.DistanceToPoint(foeDrone.Position()) <= AutoScanDistance {
					if foeDrone.DistanceToPoint(from) <= DroneSize/2 {
						score -= 1
					}
				}
			}
			for _, m := range monsters {
				if LocationDistance(from, m.Point()) <= MonsterDistanceDetect {
					score -= 1
				}
				if LocationDistance(from, m.NextPoint()) <= MonsterDistanceDetect {
					score -= 1
				}
			}

			prevStep := 2.0
			for _, m := range g.PrevMonster {
				// if d.DistanceToPoint(m.Point()) <= prevStep*MonsterDistanceDetect {
				if LocationDistance(from, m.MultNextPoint(prevStep)) <= MonsterDistanceDetect {
					score -= 1
				}
				// }
			}
			prevStep = 3.0
			for _, m := range g.PrevPrevMonster {
				// if d.DistanceToPoint(m.Point()) <= MonsterDistanceDetect {
				if LocationDistance(from, m.MultNextPoint(prevStep)) <= MonsterDistanceDetect {
					score -= 1
				}
				// }
			}
			if NearLeft(from) {
				score -= 1
			} else if NearRight(from) {
				score -= 1
			} else if NearBottom(from) {
				score -= 1
			} else if InCorner(from) {
				score -= 1
			}
			n := &Node{
				I:        i,
				J:        j,
				Point:    from,
				Distance: int(LocationDistance(from, target)),
				Score:    score,
			}
			if i == 0 || j == 0 {
				n.End = true
			}
			if i == len(location)-1 || j == len(location[i])-1 {
				n.End = true
			}
			location[i][j] = n
			j++
		}
		i++
	}
	return location
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

func (d *Drone) MoveByLocation2(location []*Node) *Node {
	max := location[0]
	dist := []*Node{max}
	for _, p := range location {
		if max.Score < p.Score {
			max = p
			dist = []*Node{max}
		} else if max.Score == p.Score {
			dist = append(dist, p)
		}
	}
	if d.NearMonster {
		minSteps := dist[0]
		distSteps := []*Node{minSteps}
		for _, p := range dist {
			if minSteps.Steps > p.Steps {
				minSteps = p
				distSteps = []*Node{minSteps}
			} else if minSteps.Steps == p.Steps {
				distSteps = append(distSteps, p)
			}
		}
		min := distSteps[0]
		for _, dd := range distSteps {
			if min.Distance > dd.Distance {
				min = dd
			}
		}
		return min
	}
	min := dist[0]
	for _, dd := range dist {
		if min.Distance > dd.Distance {
			min = dd
		}
	}
	return min
}

func (d *Drone) MoveByLocation(location [][]*Node, parent *Node) [][]*Node {
	if parent == nil {
		i, j := 0, 0
		if len(location)%2 == 0 {
			i = len(location) / 2
		} else {
			i = len(location)/2 + 1
		}
		if len(location[i])%2 == 0 {
			j = len(location[i]) / 2
		} else {
			j = len(location[i])/2 + 1
		}
		location[i][j].Used = true
		location = d.MoveByLocation(location, location[i][j])

		fitNode := location[0][0]
		scoreNodes := []*Node{}
		for i, nn := range location {
			if i == 0 || i == len(location)-1 {
				for _, n := range nn {
					if n.Score == fitNode.Score {
						scoreNodes = append(scoreNodes, n)
					} else if n.Score > fitNode.Score {
						fitNode = n
						scoreNodes = []*Node{n}
					}
				}
			} else {
				n := nn[0]
				if n.Score == fitNode.Score {
					scoreNodes = append(scoreNodes, n)
				} else if n.Score > fitNode.Score {
					fitNode = n
					scoreNodes = []*Node{n}
				}
				n = nn[len(nn)-1]
				if n.Score == fitNode.Score {
					scoreNodes = append(scoreNodes, n)
				} else if n.Score > fitNode.Score {
					fitNode = n
					scoreNodes = []*Node{n}
				}
			}
		}
		fitNode = scoreNodes[0]
		for _, n := range scoreNodes {
			if fitNode.Distance > n.Distance {
				fitNode = n
			}
		}
		fmt.Fprintf(os.Stderr, "%d MOVE %d %d %s\n", d.ID, fitNode.X, fitNode.Y, d.Light())
		fmt.Printf("MOVE %d %d %s\n", fitNode.X, fitNode.Y, d.Light())
		return location
	}

	i, j := parent.I, parent.J
	var touched bool
	for _, move := range moves {
		i = parent.I + move[0]
		if i < 0 {
			i = 0
		}
		if i >= len(location) {
			i = len(location) - 1
		}
		j = parent.J + move[1]
		if j < 0 {
			j = 0
		}
		if j >= len(location[i]) {
			j = len(location[i]) - 1
		}
		if location[i][j].Used {
			continue
		}
		location[i][j].Used = true
		location[i][j].Score += parent.Score
		touched = true
	}
	if !touched {
		return location
	}

	for _, move := range moves {
		i = parent.I + move[0]
		if i < 0 {
			i = 0
		}
		if i >= len(location) {
			i = len(location) - 1
		}
		j = parent.J + move[1]
		if j < 0 {
			j = 0
		}
		if j >= len(location[i]) {
			j = len(location[i]) - 1
		}
		location = d.MoveByLocation(location, location[i][j])
	}
	return location
}

func (d *Drone) MoveByVertex(start *Vertex) {
	scoreVertex := make([]*Vertex, 0)
	BFS(start, func(v *Vertex) {
		if v.Node.End {
			scoreVertex = append(scoreVertex, v)
		}
	})
	maxScore := scoreVertex[0].Node.Score
	distScore := []*Node{scoreVertex[0].Node}
	for _, v := range scoreVertex {
		if maxScore < v.Node.Score {
			maxScore = v.Node.Score
			distScore = []*Node{v.Node}
		} else if maxScore == v.Node.Score {
			distScore = append(distScore, v.Node)
		}
	}

	fitNode := distScore[0]
	for _, d := range distScore {
		if fitNode.Distance > d.Distance {
			fitNode = d
		}
	}

	fmt.Fprintf(os.Stderr, "%d (%d) MOVE %d %d %s\n", d.ID, len(scoreVertex), fitNode.X, fitNode.Y, d.Light())
	fmt.Printf("MOVE %d %d %s\n", fitNode.X, fitNode.Y, d.Light())
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
	CreatureCount   int
	Creatures       []*GameCreature
	MapCreatures    map[int]*GameCreature
	PrevMonster     []Creature
	PrevPrevMonster []Creature

	Resurface       map[int]Point
	DroneTarget     map[int]int
	TargetCreatures map[int]struct{}

	CreaturesTouched map[int]map[int]struct{}

	MapDroneLigthCount map[string]map[int]int

	DroneQueue         map[int][]Point
	DorneNextLightTick map[int]int

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

var moves2 = [][]int{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {0, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

var moves = [][]int{
	{-1, 0}, {0, -1}, {1, 0}, {0, 1},
}

type Vertex struct {
	ID     Point
	Node   *Node
	Parent *Vertex

	Vertices map[Point]*Vertex
}

func NewVertex(node *Node) *Vertex {
	return &Vertex{
		ID:       node.Point,
		Node:     node,
		Vertices: make(map[Point]*Vertex, 0),
	}
}

func FillLocation(i, j int, nodes [][]*Node, used map[Point]struct{}) []*Node {
	fmt.Fprintf(os.Stderr, "fill position %d %d\n", i, j)

	if i < 0 || j < 0 {
		return nil
	}
	if i >= len(nodes) || j >= len(nodes[i]) {
		return nil
	}

	if used == nil {
		used = make(map[Point]struct{})
	}
	border := make([]*Node, 0)
	// first
	for _, b := range nodes[0] {
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}
	// last
	for _, b := range nodes[len(nodes)-1] {
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}

	for j := 0; j < len(nodes); j++ {
		b := nodes[j][0]
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}
	for j := 0; j < len(nodes); j++ {
		b := nodes[j][len(nodes[j])-1]
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}

	var f func(x, y int, target *Node, parent *Node, mark map[Point]struct{}) *Node

	f = func(x, y int, target *Node, parent *Node, mark map[Point]struct{}) *Node {
		if mark == nil {
			mark = map[Point]struct{}{}
		}
		result := nodes[x][y]
		if parent.I != result.I && parent.J != result.J {
			result.Parent = parent
		}
		min := LocationDistance(Point{result.I, result.J}, Point{target.I, target.J})
		for _, move := range moves2 {
			i := x + move[0]
			j := y + move[1]
			if i < 0 || j < 0 {
				continue
			}
			if i >= len(nodes) || j >= len(nodes[i]) {
				continue
			}
			node := nodes[i][j]
			if _, ok := mark[node.Point]; ok {
				continue
			}
			node.Parent = parent
			newMin := LocationDistance(Point{node.I, node.J}, Point{target.I, target.J})
			if min > newMin {
				min = newMin
				result = node
			}
			mark[node.Point] = struct{}{}
		}

		if result.I == target.I && result.J == target.J {
			target.Parent = parent
			return target
		}
		result = f(result.I, result.J, target, result, mark)
		return result
	}

	// markScored := map[Point]struct{}{}
	for m, b := range border {
		parent := nodes[i][j]
		newB := f(i, j, b, parent, nil)
		if newB == nil {
			continue
		}
		border[m] = newB

		next := newB
		// score := next.Score
		// if next.Parent != nil {
		// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, next.Parent.I, next.Parent.J)
		// } else {
		// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, parent.I, parent.J)
		// }
		path := []*Node{next}
		for {
			parent := next.Parent
			if parent == nil {
				break
			}
			path = append(path, parent)
			// parent.Score += score
			// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, parent.I, parent.J)
			if parent.I == i && parent.J == j {
				break
			}
			next = parent
		}
		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}
		score := 0
		odd := 0
		if len(nodes)%2 == 0 {
			odd = 1
		}
		for i, p := range path {
			score += p.Score
			// fmt.Fprintf(os.Stderr, "(%d:%d:%d)\n", p.I, p.J, score)
			if i == len(path)-1 {
				if p.X == 0 || p.Y == 0 {
					p.Score = score - odd
					p.Steps = i + odd
				} else {
					p.Score = score
					p.Steps = i
				}
			}
		}
	}
	return border
}

func NewGraph(i, j int, nodes [][]*Node, used map[Point]struct{}, parent *Vertex) *Vertex {

	if i < 0 || j < 0 {
		return nil
	}
	if i >= len(nodes) || j >= len(nodes[i]) {
		return nil
	}

	start := nodes[i][j]
	if _, ok := used[start.Point]; ok {
		return nil
	}
	start.Used = true
	root := &Vertex{
		ID:     start.Point,
		Node:   start,
		Parent: parent,
	}

	if used == nil {
		used = map[Point]struct{}{}
	}
	used[root.ID] = struct{}{}
	// fmt.Fprintln(os.Stderr, i, j)
	vertices := []*Vertex{}
	for _, move := range moves {
		i := i + move[0]
		j := j + move[1]
		if i < 0 || j < 0 {
			continue
		}
		if i >= len(nodes) || j >= len(nodes[i]) {
			continue
		}
		node := nodes[i][j]
		if node.Used {
			continue
		}
		node.Used = true
		// fmt.Fprintf(os.Stderr, "%v:%d:%d|", move, i, j)
		v := NewVertex(node)
		if v == nil {
			continue
		}
		vertices = append(vertices, v)
	}

	for _, v := range vertices {
		v.Parent = root
		if root.Vertices == nil {
			root.Vertices = map[Point]*Vertex{}
		}

		v := NewGraph(v.Node.I, v.Node.J, nodes, used, root)
		if v != nil {
			root.Vertices[v.ID] = v
		}
		used[v.ID] = struct{}{}
	}

	return root
}

func DebugVertex(v *Vertex) {
	fmt.Fprintf(os.Stderr, "%d:%d:%d:%d\n", v.ID.X, v.ID.Y, v.Node.Score, v.Node.Distance)
	for _, k := range v.Vertices {
		fmt.Fprintf(os.Stderr, "(%d:%d:%d:%d)|", k.ID.X, k.ID.Y, k.Node.Score, v.Node.Distance)
	}
	fmt.Fprintln(os.Stderr)
	for _, vv := range v.Vertices {
		DebugVertex(vv)
	}
}

// create a node that holds the graphs vertex as data
type node struct {
	v    *Vertex
	next *node
}

// create a queue data structure
type queue struct {
	head *node
	tail *node
}

// enqueue adds a new node to the tail of the queue
func (q *queue) enqueue(v *Vertex) {
	n := &node{v: v}

	// if the queue is empty, set the head and tail as the node value
	if q.tail == nil {
		q.head = n
		q.tail = n
		return
	}

	q.tail.next = n
	q.tail = n
}

// dequeue removes the head from the queue and returns it
func (q *queue) dequeue() *Vertex {
	n := q.head
	// return nil, if head is empty
	if n == nil {
		return nil
	}

	q.head = q.head.next

	// if there wasn't any next node, that
	// means the queue is empty, and the tail
	// should be set to nil
	if q.head == nil {
		q.tail = nil
	}

	return n.v
}

func BFS(startVertex *Vertex, visitCb func(*Vertex)) {
	// initialize queue and visited vertices map
	vertexQueue := &queue{}
	visitedVertices := map[Point]struct{}{}

	currentVertex := startVertex
	// start a continuous loop
	for {
		// visit the current node
		visitCb(currentVertex)
		visitedVertices[currentVertex.ID] = struct{}{}

		// for each neighboring vertex, push it to the queue
		// if it isn't already visited
		for _, v := range currentVertex.Vertices {
			if _, ok := visitedVertices[v.ID]; !ok {
				vertexQueue.enqueue(v)
			}
		}

		// change the current vertex to the next one
		// in the queue
		currentVertex = vertexQueue.dequeue()
		// if the queue is empty, break out of the loop
		if currentVertex == nil {
			break
		}
	}
}

func TestNewGraph3x3(t *testing.T) {
	src := [][]*Node{}
	max := 4

	for i := 0; i < max; i++ {
		src = append(src, make([]*Node, 0))
		for j := 0; j < max; j++ {
			src[i] = append(src[i], &Node{
				Point:    Point{X: i, Y: j},
				I:        i,
				J:        j,
				Score:    -1 * rand.Intn(10),
				Distance: -i - j,
			},
			)
		}
	}

	i, j := max/2, max/2
	if max%2 == 0 {
		i, j = i-1, j-1
	}
	t.Logf("%d:%d", i, j)
	DebugLocation(src, 0, Point{X: i, Y: j})
	path := FillLocation(i, j, src, nil)
	DebugLocation(src, 0, Point{X: i, Y: j})
	drone := Drone{}
	p := drone.MoveByLocation2(path)
	for _, p := range path {
		t.Logf("%d:%d:%d:%d", p.I, p.J, p.Score, p.Steps)
	}
	t.Logf("%d:%d:%d", p.I, p.J, p.Score)
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
				posX := drone.X
				deltaX := 0
				if drone.X < MaxPosistionX/2 {
					if drone.ID%2 == 0 {
						deltaX = 2000
					} else {
						deltaX = 1000
					}
				} else {
					if drone.ID%2 == 0 {
						deltaX = -1000
					} else {
						deltaX = -2000
					}
				}
				// if drone.X < MaxPosistionX/2 {
				// posX = 2000
				// } else {
				// posX = 8000
				// }
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
			}
		})

		leftDrone, rightDrone := s.MyDrones[0].ID, s.MyDrones[1].ID
		if s.MyDrones[0].X > s.MyDrones[1].X {
			leftDrone, rightDrone = s.MyDrones[1].ID, s.MyDrones[0].ID
		}
		if len(game.DroneQueue) < 12 {
			hashTarget := map[int]struct{}{}
			for _, r := range s.Radar {
				creature := game.GetCreature(r.CreatureID)
				drone := s.GetDrone(r.DroneID)

				if _, ok := s.MyCreatures[r.CreatureID]; ok {
					continue
				}
				var scanned bool
				for _, k := range s.MyDrones {
					if _, ok := s.DroneScnas[k.ID][r.CreatureID]; ok {
						for _, drone := range s.MyDrones {
							if v, ok := game.DroneTarget[drone.ID]; ok && r.CreatureID == v {
								delete(game.DroneTarget, drone.ID)
							}
						}
						scanned = true
						break
					}
				}
				if _, ok := hashTarget[r.CreatureID]; ok {
					continue
				}
				if game.OnDroneDepth(creature, drone) {
					if scanned {
						fmt.Fprintln(os.Stderr, "skip target", creature.ID, drone.ID)
						continue
					}
					switch r.Radar {
					case RadarTL:
						if v, ok := s.DroneCreatureRadar[rightDrone]; ok {
							if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
								firstPoint := game.FirstCommand(*drone)
								// to resurface
								if firstPoint.Y < MaxPosistionY/2 {
									game.DroneTarget[leftDrone] = r.CreatureID
									hashTarget[r.CreatureID] = struct{}{}
								}
							}
						}
					case RadarTR:
						if v, ok := s.DroneCreatureRadar[leftDrone]; ok {
							if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
								firstPoint := game.FirstCommand(*drone)
								// to resurface
								if firstPoint.Y < MaxPosistionY/2 {
									game.DroneTarget[rightDrone] = r.CreatureID
									hashTarget[r.CreatureID] = struct{}{}
								}
							}
						}
					case RadarBL:
						if v, ok := s.DroneCreatureRadar[rightDrone]; ok {
							if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
								firstPoint := game.FirstCommand(*drone)
								// to bottom
								if firstPoint.Y > MaxPosistionY/2 {
									game.DroneTarget[leftDrone] = r.CreatureID
									hashTarget[r.CreatureID] = struct{}{}
								}
							}
						}
					case RadarBR:
						if v, ok := s.DroneCreatureRadar[leftDrone]; ok {
							if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
								firstPoint := game.FirstCommand(*drone)
								// to bottom
								if firstPoint.Y > MaxPosistionY/2 {
									game.DroneTarget[rightDrone] = r.CreatureID
									hashTarget[r.CreatureID] = struct{}{}
								}
							}
						}
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "target: %v\nscanned:%d\n", game.DroneTarget, s.DroneScnas)

		for i := range s.MyDrones {
			drone := s.MyDrones[i]
			if drone.IsEmergency() {
				game.RemoveDroneTarget(drone)
				fmt.Fprintln(os.Stderr, drone.ID, "emergency", s.DroneScnas[drone.ID])
				drone.Wait()
				continue
			}
			for _, s := range s.Creatures {
				gs := game.GetCreature(s.ID)
				if gs.Type < 0 && drone.DistanceToPoint(s.Point()) < AutoScanDistance {
					drone.NearMonster = true
				}
			}

			drone.TurnLight(game)
			drone.SolveRadarRadius(game, s.MapRadar[drone.ID])
			drone.DebugRadarRadius()

			newPoint := game.FirstCommand(drone)
			if drone.DistanceToPoint(newPoint) < SurfaceDistance {
				newPoint = game.PopCommand(drone)
				if drone.Y > MaxPosistionY/2 {
					game.DroneQueue[drone.ID][0].X = drone.X
					newPoint.X = drone.X
				}
			}
			m := drone.Solve(game, s, s.MapRadar[drone.ID], newPoint)
			if drone.ID == 0 || drone.ID == 3 {
				// DebugLocation(m, drone.ID, newPoint)
			}
			//v := drone.SolveToGraph(game, s, m, newPoint)
			//BFS(v, func(v *Vertex) {
			//	// fmt.Fprintf(os.Stderr, ">>(%d:%d:%d)\n", v.ID.X, v.ID.Y, len(v.Vertices))
			//	for k := range v.Vertices {
			//		v.Vertices[k].Node.Score += v.Node.Score
			//	}
			//})
			path := drone.SolveFillLocation(m, nil)
			if drone.ID == 0 || drone.ID == 3 {
				DebugLocation(m, drone.ID, newPoint)
			}
			// DebugVertex(v)
			DebugPath(path)

			// drone.MoveByVertex(v)
			// m = drone.MoveByLocation(m, nil)
			p := drone.MoveByLocation2(path)
			drone.Move(p.Point)

			// DebugLocation(m, drone.ID, newPoint)
		}
		if s.CreaturesAllScannedOrSaved() {
			for _, drone := range s.MyDrones {
				newPoint := game.FirstCommand(drone)
				if newPoint.Y != ResurfaceDistance && len(game.DroneQueue[drone.ID]) >= 1 {
					game.DroneQueue[drone.ID][0] = Point{X: drone.X, Y: ResurfaceDistance}
				}
			}
		}
		if len(game.PrevMonster) > 0 {
			game.PrevPrevMonster = make([]Creature, 0)
			for _, p := range game.PrevMonster {
				game.PrevPrevMonster = append(game.PrevPrevMonster, p)
			}

		}
		game.PrevMonster = make([]Creature, 0)
		for _, s := range s.Creatures {
			gs := game.GetCreature(s.ID)
			if gs.Type < 0 {
				game.PrevMonster = append(game.PrevMonster, s)
			}
		}
	}
}

type Node struct {
	Point
	I      int
	J      int
	Parent *Node
	Steps  int

	CreaturesTypes []*GameCreature
	Score          int
	Radar          string
	Distance       int

	Drone    *Drone
	FoeDrone *Drone
	Creature *Creature
	Used     bool
	End      bool
}

func (n *Node) StringID() string {
	return fmt.Sprintf("%d:%d", n.X, n.Y)
}

func DebugLocation(location [][]*Node, droneID int, target Point) {
	fmt.Fprintln(os.Stderr, "debug location", droneID, target)
	for _, nn := range location {
		for _, n := range nn {
			fmt.Fprintf(os.Stderr, "%d.%d.%d.%d.%d|", n.X, n.Y, n.Distance, n.Score, n.Steps)
		}
		fmt.Fprintln(os.Stderr)
	}
}

func NearestNode(nodes []Node, pos Point) Node {
	min := nodes[0]
	minDist := LocationDistance(min.Point, pos)
	for _, n := range nodes {
		if newMinDist := LocationDistance(n.Point, pos); newMinDist < minDist {
			minDist = newMinDist
			min = n
		}
	}
	return min
}

func LocationDistance(from Point, to Point) float64 {
	return math.Sqrt(math.Pow(float64(to.X-from.X), 2) + math.Pow(float64(to.Y-from.Y), 2))
}

func LeftSide(p Point) Point {
	return Point{X: 0, Y: p.Y}
}

func RightSide(p Point) Point {
	return Point{X: MaxPosistionX - AutoScanDistance, Y: p.Y}
}

func BottomSide(p Point) Point {
	return Point{X: p.X, Y: MaxPosistionY - AutoScanDistance}
}

func NearLeft(p Point) bool {
	side := LeftSide(p)
	return math.Abs(LocationDistance(side, p)) < DroneSize/2
}

func NearRight(p Point) bool {
	side := RightSide(p)
	return math.Abs(LocationDistance(side, p)) < DroneSize/2
}

func NearBottom(p Point) bool {
	side := BottomSide(p)
	return math.Abs(LocationDistance(side, p)) < DroneSize/2
}

func InCorner(p Point) bool {
	return (NearRight(p) && NearBottom(p)) || (NearLeft(p) && NearBottom(p))
}

func DebugPath(path []*Node) {
	for _, p := range path {
		fmt.Fprintf(os.Stderr, "(%d:%d:%d:%d)", p.X, p.Y, p.Score, p.Steps)
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

func RadarToScore(radar string, p Point, drone *Drone) int {
	switch radar {
	case RadarTL:
		if p.X < drone.X && p.Y < drone.Y {
			return 1
		}
	case RadarTR:
		if p.X > drone.X && p.Y < drone.Y {
			return 1
		}
	case RadarBL:
		if p.X < drone.X && p.Y > drone.Y {
			return 1
		}
	case RadarBR:
		if p.X > drone.X && p.Y > drone.Y {
			return 1
		}
	}
	return 0
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

	RadarBlipCount     int
	Radar              []Radar
	MapRadar           map[int][]Radar
	DroneCreatureRadar map[int]map[int]string
	AvailableCreatures map[int]struct{}
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

func (s *State) CreaturesAllScannedOrSaved() bool {
	hash := map[int]struct{}{}
	for _, d := range s.MyDrones {
		if v, ok := s.DroneScnas[d.ID]; ok {
			for k := range v {
				hash[k] = struct{}{}
			}
		}
	}
	allScanned := true
	for k := range s.MyCreatures {
		if _, ok := s.AvailableCreatures[k]; !ok {
			allScanned = false
			break
		}
	}
	fmt.Fprintf(os.Stderr, "%d %d %d\n", len(s.AvailableCreatures), len(hash), len(s.MyCreatures))
	if !allScanned {
		return false
	}
	return len(s.AvailableCreatures) <= (len(hash) + len(s.MyCreatures))
}

func NewState(g *GameState) *State {
	s := &State{
		MyCreatures:        make(map[int]struct{}, 0),
		FoeCreatures:       make([]int, 0),
		MyDrones:           make([]Drone, 0),
		FoeDrones:          make([]Drone, 0),
		DroneScnas:         make(map[int]map[int]struct{}, 0),
		Creatures:          make([]Creature, 0),
		Radar:              make([]Radar, 0),
		MapRadar:           make(map[int][]Radar, 0),
		DroneCreatureRadar: make(map[int]map[int]string, 0),
		AvailableCreatures: make(map[int]struct{}, 0),
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
		if v, ok := s.DroneCreatureRadar[r.DroneID]; !ok {
			s.DroneCreatureRadar[r.DroneID] = map[int]string{r.CreatureID: r.Radar}
		} else {
			v[r.CreatureID] = r.Radar
			s.DroneCreatureRadar[r.DroneID] = v
		}

		if g.GetCreature(r.CreatureID).Type >= 0 {
			s.AvailableCreatures[r.CreatureID] = struct{}{}
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
