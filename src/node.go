package main

import (
	"fmt"
	"math"
	"os"
)

type Node struct {
	Point

	CreaturesTypes []*GameCreature
	Score          int
	Radar          string
	Distance       int

	Drone    *Drone
	FoeDrone *Drone
	Creature *Creature
}

func DebugLocation(location [][]Node, droneID int, target Point) {
	fmt.Fprintln(os.Stderr, "debug location", droneID, target)
	for _, nn := range location {
		for _, n := range nn {
			fmt.Fprintf(os.Stderr, "%d.%d.%d.%d|", n.X, n.Y, n.Distance, n.Score)
		}
		fmt.Fprintln(os.Stderr)
	}
}

func LocationDistance(from Point, to Point) float64 {
	return math.Sqrt(math.Pow(float64(to.X-from.X), 2) + math.Pow(float64(to.Y-from.Y), 2))
}
