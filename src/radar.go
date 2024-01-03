package main

import (
	"fmt"
)

type Radar struct {
	DroneID    int
	CreatureID int
	Radar      string
}

func AngelToRadar(angle int) string {
	if angle <= 90 {
		return RadarTR
	} else if angle <= 180 {
		return RadarTL
	} else if angle <= 270 {
		return RadarBL
	} else {
		return RadarBR
	}
}

func RadarToDirection(radar string) (x int, y int) {
	switch radar {
	case RadarTL:
		return -1, -1
	case RadarTR:
		return 1, -1
	case RadarBR:
		return 1, 1
	default:
		return -1, 1
	}
}

func NewRadar() Radar {
	r := Radar{}
	fmt.Scan(&r.DroneID, &r.CreatureID, &r.Radar)
	return r
}
