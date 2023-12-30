package main

import (
	"fmt"
)

type Radar struct {
	DroneID    int
	CreatureID int
	Radar      string
}

func NewRadar() Radar {
	r := Radar{}
	fmt.Scan(&r.DroneID, &r.CreatureID, &r.Radar)
	return r
}
