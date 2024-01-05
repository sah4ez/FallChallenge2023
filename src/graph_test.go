package main

import (
	"math/rand"
	"testing"
)

func TestNewGraph3x3(t *testing.T) {
	src := [][]*Node{}
	max := 5

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
