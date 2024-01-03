package main

type Node struct {
	Point

	CreaturesTypes []*GameCreature
	Score          int
	Radar          string

	Drone *Drone
	FoeDrone *Drone
	Creature *Creature
}
