package main

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
