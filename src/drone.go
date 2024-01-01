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

func (d *Drone) TurnLight(g *GameState) {
	cnt := g.GetCoutLights(d)
	if d.Battery > LightBattary && cnt > 0 {
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

func (d *Drone) MoveToRadarByMonster(radar string, msg ...string) {
	p := ShiftByRadarByMonster(radar, d.X, d.Y)
	d.Move(p, msg...)
}

func (d *Drone) RandMove() {
	x := rand.Intn(MaxRandStep-MinRandStep) + MinRandStep
	y := rand.Intn(MaxRandStep-MinRandStep) + MinRandStep
	d.Move(Point{X: d.X + x, Y: d.Y + y}, "random")
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
	d := Drone{}
	fmt.Scan(&d.ID, &d.X, &d.Y, &d.Emergency, &d.Battery)
	fmt.Fprintln(os.Stderr, "scan drone", d.ID, d.X, d.Y, d.Emergency, d.Battery)
	return d
}
