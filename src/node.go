package main

import (
	"fmt"
	"math"
	"os"
)

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
			fmt.Fprintf(os.Stderr, "%d.%d.%d.%d|", n.X, n.Y, n.Distance, n.Score)
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
