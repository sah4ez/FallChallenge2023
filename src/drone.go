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

func (d *Drone) Solve(g *GameState, s *State, radar []Radar, target Point) [][]Node {
	startX := d.X - AutoScanDistance
	if startX < 0 {
		startX = 0
	}
	startY := d.Y - AutoScanDistance
	if startY < 0 {
		startY = 0
	}
	endX := d.X + AutoScanDistance
	if endX >= MaxPosistionX {
		endX = MaxPosistionX
	}
	endY := d.Y + AutoScanDistance
	if endY >= MaxPosistionY {
		endY = MaxPosistionY
	}

	location := make([][]Node, int(math.Floor(float64(endX-startX)/StepScan)))
	i := 0
	monsters := []Creature{}
	for _, c := range s.Creatures {
		gc := g.GetCreature(c.ID)
		if gc.Type < 0 {
			monsters = append(monsters, c)
		}
	}
	fmt.Fprintln(os.Stderr, "monsters", len(monsters))
	for x := startX; x < endX; x += int(StepScan) {
		if i >= len(location) {
			continue
		}
		location[i] = make([]Node, int(math.Floor(float64(endY-startY)/StepScan)))
		j := 0
		for y := startY; y < endY; y += int(StepScan) {
			if j >= len(location[i]) {
				continue
			}
			from := Point{X: x, Y: y}
			score := 0
			for _, m := range monsters {
				if LocationDistance(from, m.Point()) <= AutoScanDistance {
					score -= 1
				}
				if LocationDistance(from, m.NextPoint()) <= AutoScanDistance {
					score -= 1
				}
			}
			n := Node{
				Point:    from,
				Distance: int(LocationDistance(from, target)),
				Score:    score,
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
