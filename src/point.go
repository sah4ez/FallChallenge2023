package main

import (
	"fmt"
	"os"
)

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
	origP := Point{X: x, Y: y}
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
	fmt.Fprintln(os.Stderr, "by radar", radar, p.String(), origP.String())
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
