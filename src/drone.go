package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
)

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
	// cnt := g.GetCoutLights(d)

	if d.Battery >= LightBattary && d.Y > MinLigthDepth {
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

	location := make([][]*Node, int(math.Floor(float64(endX-startX)/StepScan)))
	i := 0
	monsters := []Creature{}
	for _, c := range s.Creatures {
		gc := g.GetCreature(c.ID)
		if gc.Type < 0 && d.DistanceToPoint(c.Point()) < MaxAutoScanDistance {
			monsters = append(monsters, c)
		}
	}
	fmt.Fprintln(os.Stderr, "monsters", len(monsters), "prev", len(g.PrevMonster))
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
		location[i] = make([]*Node, int(math.Floor(float64(endY-startY)/StepScan)))
		j := 0
		for y := startY; y < endY; y += int(StepScan) {
			if j >= len(location[i]) {
				continue
			}
			from := Point{X: x, Y: y}
			score := 0
			if d.NearMonster && false {
				nearest := NearestNode(d.radiusPoint, from)
				for _, c := range nearest.CreaturesTypes {
					if c.Type < 0 {
						score -= 1
					}
					// else {
					// score += 1
					// }
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

			if len(monsters) == 0 {
				prevStep := 2.0
				for _, m := range g.PrevMonster {
					if d.DistanceToPoint(m.Point()) <= prevStep*MonsterDistanceDetect {
						if LocationDistance(from, m.MultNextPoint(prevStep)) <= MonsterDistanceDetect {
							score -= 1
						}
					}
				}
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
